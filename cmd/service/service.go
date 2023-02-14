// TODO
package service

import "context"

type Storage interface {
	Insert(ctx context.Context, data any)
	Get(ctx context.Context, key string)
}

type Uploader interface {
	Upload()
	Get()
}

type LinkDeps struct {
	storage  Storage
	uploader Uploader
}

func NewLinkDeps(storage Storage, uploader Uploader) *LinkDeps {
	return &LinkDeps{
		storage:  storage,
		uploader: uploader,
	}
}

func InsertLink(ctx context.Context) {

}
