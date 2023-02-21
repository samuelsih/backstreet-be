package service

import (
	"backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"backstreetlinkv2/cmd/repo"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
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
	Get(ctx context.Context, filename string, wr io.Writer) (repo.FileStat, error)
}

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, val []byte) error
}

type Deps struct {
	storage  Storage
	uploader Uploader
	cache    Cache
}

func NewLinkDeps(storage Storage, uploader Uploader, cache Cache) *Deps {
	return &Deps{
		storage:  storage,
		uploader: uploader,
		cache:    cache,
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

	defer func() {
		marshalled, err := json.Marshal(data)
		if err != nil {
			log.Warn().Err(err).Msg("cant marshal InsertLink")
			return
		}

		if err := d.cache.Set(data.Alias, marshalled); err != nil {
			log.Warn().Err(err).Msg("cant store to cache in InsertLink")
		}
	}()

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

	var err error

	err = d.uploader.Upload(ctx, data.Alias, data.RawFile)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	err = d.storage.Insert(ctx, data.Alias, data)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	defer func() {
		marshalled, err := json.Marshal(data)
		if err != nil {
			log.Warn().Err(err).Msg("cant marshal InsertFile")
			return
		}

		if err := d.cache.Set(data.Alias, marshalled); err != nil {
			log.Warn().Err(err).Msg("cant store to cache in InsertFile")
		}
	}()

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

	resultFromMemory, err := d.cache.Get(key)
	if (err != nil) && !errors.Is(err, repo.CacheNotFound) {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	if errors.Is(err, repo.CacheNotFound) {
		result, err := d.storage.Get(ctx, key)
		if err != nil {
			out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
			return out
		}

		out.Response = result

		out.SetOK()
		return out
	}

	if err := json.Unmarshal(resultFromMemory, &out.Response); err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, `can't get your data`))
		return out
	}

	out.SetOK()
	return out
}

type DownloadFileOutput struct {
	CommonResponse
	ContentDisposition string        `json:"-"`
	ContentType        string        `json:"-"`
	ContentLength      string        `json:"-"`
	File               io.ReadWriter `json:"-"`
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

	out.File = bytes.NewBuffer([]byte{})

	fs, err := d.uploader.Get(ctx, record.Alias, out.File)
	if err != nil {
		out.SetErr(helper.E(op, helper.GetKind(err), err, err.Error()))
		return out
	}

	out.ContentType = fs.ContentType
	out.ContentLength = strconv.FormatInt(fs.ContentLength, 10)
	out.ContentDisposition = fmt.Sprintf("attachment; filename=\"%s\"", record.Filename)

	out.SetOK()
	return out
}
