package helper

import (
	"backstreetlinkv2/cmd/model"
	"github.com/go-playground/validator/v10"
)

type Request interface {
	model.ShortenRequest | model.ShortenFileRequest
}

func ValidateStruct[T Request](data T) error {
	validate := validator.New()

	err := validate.Struct(data)

	if err != nil {
		return err
	}

	return nil
}
