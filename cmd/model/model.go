package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
)

const (
	TypeLink = "LINK"
	TypeFile = "FILE"
)

var (
	ByteAssertionErr = errors.New("byte assertion failed")
)

type Shorten struct {
	Alias      string `json:"alias"`
	RedirectTo string `json:"redirect_to"`
	Type       string `json:"type"`
	Filename   string `json:"data_source"`
}

type ShortenRequest struct {
	Alias      string `json:"alias" validate:"required,min=5,max=30,alphanum"`
	Type       string `json:"type" validate:"eq=LINK"`
	RedirectTo string `json:"redirect_to" validate:"url"`
}

type ShortenFileRequest struct {
	Alias    string        `json:"alias" validate:"required,min=5,max=30,alphanum"`
	Filename string        `json:"-"`
	Type     string        `json:"type" validate:"oneof='FILE'"`
	RawFile  io.ReadCloser `json:"-"`
}

type ShortenResponse struct {
	Type       string `json:"type"`
	Alias      string `json:"alias"`
	RedirectTo string `json:"redirect_to"`
	Filename   string `json:"filename"`
}

func (s ShortenResponse) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ShortenResponse) Scan(val any) error {
	var b []byte

	switch v := val.(type) {
	case string:
		b = []byte(v)

	case []byte:
		b = v

	default:
		return ByteAssertionErr
	}

	return json.Unmarshal(b, &s)
}
