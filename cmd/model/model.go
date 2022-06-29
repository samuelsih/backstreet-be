package model

const (
	TypeLink = "LINK"
	TypeFile = "FILE"
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
	Alias          string `json:"alias" bson:"_id" validate:"required,min=5,max=30,alphanum"`
	Filename       string `json:"-" bson:"filename"`        //nama file asli
	FilenameSource string `json:"-" bson:"filename_source"` //nama file di cloud
	Type           string `json:"type" bson:"type" validate:"oneof='FILE'"`
}

type ShortenResponse struct {
	Alias          string `json:"alias" bson:"_id"`
	RedirectTo     string `json:"redirect_to" bson:"redirect_to"`
	Type           string `json:"type" bson:"type"`
	Filename       string `json:"filename" bson:"filename"` //nama file asli
	FilenameSource string `json:"-" bson:"filename_source"` //nama file di cloud
}
