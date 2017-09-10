package server

import (
	"time"
)

type Participant struct {
	Uuid      string
	Last      *Image
	Connected *Participant
}

type Image struct {
	Path string
	Date time.Time
	Seen bool
}

type Configuration struct {
	Root, Key string
}

// API
type GenericResponse struct {
	Error   string      `json:"error,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

type GetImageResponse struct {
	Url  string    `json:"url"`
	Date time.Time `json:"date"`
}

type PostImageRequest struct {
	Mime string `json:"mime"`
	Data []byte `json:"data"`
}
