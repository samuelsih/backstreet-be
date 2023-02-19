package service

import (
	"backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strconv"
)

var (
	WrongTypeErr = errors.New("invalid request")
)

const (
	CantProcessRequest = "can't process your request"
)

type Storage interface {
	Insert(ctx context.Context, key string, data any) error
	Get(ctx context.Context, key string) (model.ShortenResponse, error)
}

type Uploader interface {
	Upload(ctx context.Context, filename string, file io.ReadCloser) error
	Get(ctx context.Context, filename string, file io.WriterAt) error
}

type Deps struct {
	storage  Storage
	uploader Uploader
}

func NewLinkDeps(storage Storage, uploader Uploader) *Deps {
	return &Deps{
		storage:  storage,
		uploader: uploader,
	}
}

type InsertLinkOutput struct {
	CommonResponse
	Alias      string `json:"alias"`
	Type       string `json:"type"`
	RedirectTo string `json:"redirect_to"`
}

func (d *Deps) InsertLink(ctx context.Context, data model.ShortenRequest) InsertLinkOutput {
	const op = helper.Op("InsertLink")
	var out InsertLinkOutput

	if data.Type != model.TypeLink {
		out.SetErr(helper.E(op, helper.KindBadRequest, WrongTypeErr, WrongTypeErr.Error()))
		return out
	}

	err := d.storage.Insert(ctx, data.Alias, data)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	out.Alias = data.Alias
	out.Type = data.Type
	out.RedirectTo = data.RedirectTo

	out.SetOK()
	return out
}

type InsertFileOutput struct {
	CommonResponse
	Alias    string `json:"alias"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
}

func (d *Deps) InsertFile(ctx context.Context, data model.ShortenFileRequest) InsertFileOutput {
	const op = helper.Op("InsertFile")
	var out InsertFileOutput

	if data.Type != model.TypeFile {
		out.SetErr(helper.E(op, helper.KindBadRequest, WrongTypeErr, WrongTypeErr.Error()))
		return out
	}

	file, err := os.Create(data.Filename)
	if err != nil {
		out.SetErr(err)
	}

	defer file.Close()

	err = d.uploader.Upload(ctx, data.Filename, data.RawFile)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	err = d.storage.Insert(ctx, data.Alias, data)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	out.Alias = data.Alias
	out.Type = data.Type
	out.Filename = data.Filename

	out.SetOK()
	return out
}

type FindOutput struct {
	CommonResponse
	Response model.ShortenResponse `json:"response"`
}

func (d *Deps) Find(ctx context.Context, key string) FindOutput {
	const op = helper.Op("Find")
	var out FindOutput

	result, err := d.storage.Get(ctx, key)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	out.Response = result

	out.SetOK()
	return out
}

type DownloadFileOutput struct {
	CommonResponse
	ContentDisposition string    `json:"-"`
	ContentType        string    `json:"-"`
	ContentLength      string    `json:"-"`
	File               io.Reader `json:"-"`
}

func (d *Deps) DownloadFile(ctx context.Context, key string) DownloadFileOutput {
	const op = helper.Op("DownloadFile")
	var out DownloadFileOutput

	record, err := d.storage.Get(ctx, key)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	if record.Type != model.TypeFile {
		out.SetErr(helper.E(op, helper.KindBadRequest, WrongTypeErr, WrongTypeErr.Error()))
		return out
	}

	file, err := os.Open(record.Filename)
	if err != nil {
		out.SetErr(helper.E(op, helper.KindUnexpected, err, CantProcessRequest))
		return out
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Err(err)
		}
	}()

	err = d.uploader.Get(ctx, record.Filename, file)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	fileInfo, err := file.Stat()
	if err != nil {
		out.SetErr(helper.E(op, helper.KindUnexpected, err, CantProcessRequest))
		return out
	}

	out.ContentLength = strconv.FormatInt(fileInfo.Size(), 10)
	out.ContentType = "multipart/form-data"
	out.ContentDisposition = fmt.Sprintf(`attachment; filename="%s"`, fileInfo.Name())

	out.SetOK()
	return out
}
