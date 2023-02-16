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
	Alias      string `json:"alias" bson:"_id"`
	RedirectTo string `json:"redirect_to" bson:"redirect_to,omitempty"`
	Type       string `json:"type" bson:"type"`
	Filename   string `json:"data_source" bson:"data_source,omitempty"`
}

type ShortenRequest struct {
	Alias      string `json:"alias" bson:"_id" validate:"required,min=5,max=30,alphanum"`
	Type       string `json:"type" bson:"type" validate:"eq=LINK"`
	RedirectTo string `json:"redirect_to" bson:"redirect_to,omitempty" validate:"url"`
}

type ShortenFileRequest struct {
	Alias    string    `json:"alias" bson:"_id" validate:"required,min=5,max=30,alphanum"`
	Filename string    `json:"-" bson:"filename"`
	Type     string    `json:"type" bson:"type" validate:"oneof='FILE'"`
	RawFile  io.Reader `json:"-"`
}

type ShortenResponse struct {
	Alias      string `json:"alias"`
	RedirectTo string `json:"redirect_to"`
	Type       string `json:"type"`
	Filename   string `json:"filename"` //nama file asli
}

func (s ShortenResponse) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ShortenResponse) Scan(val any) error {
	b, ok := val.([]byte)
	if !ok {
		return ByteAssertionErr
	}

	return json.Unmarshal(b, &s)
}
