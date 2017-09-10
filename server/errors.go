package server

import "errors"

var (
	ErrInvalidKey         = errors.New("Invalid API key supplied")
	ErrNoSlots            = errors.New("No slots left")
	ErrImageUpload        = errors.New("Image could not be uploaded")
	ErrMediaNotRecognized = errors.New("Image mime type not recognized")
)
