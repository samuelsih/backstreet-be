// TODO
package service

import (
	"backstreetlinkv2/cmd/model"
	"context"
	"io"
	"mime/multipart"
)

type Storage interface {
	Insert(ctx context.Context, key string, data any) error
	Get(ctx context.Context, key string) (model.ShortenResponse, error)
}

type Uploader interface {
	Upload(ctx context.Context, filename string, file multipart.File) error
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
}

func (d *Deps) InsertLink(ctx context.Context, data model.ShortenRequest) InsertLinkOutput {
	var resp InsertLinkOutput

	return resp
}

func (d *Deps) InsertFile(ctx context.Context) {

}

type FindOutput struct {
	CommonResponse
	Response model.ShortenResponse
}

func (d *Deps) Find(ctx context.Context, key string) FindOutput {
	var out FindOutput

	result, err := d.storage.Get(ctx, key)
	if err != nil {
		out.Set(determineErr(err), err.Error())
		return out
	}

	out.Response = result

	return out
}

type DownloadFileOutput struct {
	CommonResponse
	ContentDisposition string
	ContentType        string
	ContentLength      string
}

func (d *Deps) DownloadFile(ctx context.Context, key string) {

}

// TODO
func determineErr(err error) int {
	return 0
}
