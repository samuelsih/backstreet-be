package helper

import (
	"backstreetlinkv2/cmd/model"
	"errors"

	"github.com/go-playground/validator/v10"
)

type request interface {
	model.ShortenRequest | model.ShortenFileRequest
}

func ValidateStruct[T request](data T) (error) {
	if isZero(data) {
		return errors.New("unknown request type")
	}

	validate := validator.New()

	err := validate.Struct(data)

	if err != nil {
		return err
	}

	return nil
}

func isZero[T comparable](v T) bool {
	return v == *new(T)
}