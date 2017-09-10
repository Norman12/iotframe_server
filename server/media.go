package server

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	mimeExts = map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
	}
)

type Media struct {
	c *Configuration
}

func NewMedia(c *Configuration) *Media {
	return &Media{
		c: c,
	}
}

func (m *Media) Save(req *PostImageRequest) (string, error) {
	var ext string
	{
		if e, ok := mimeExts[req.Mime]; ok {
			ext = e
		} else {
			return "", ErrMediaNotRecognized
		}
	}

	var (
		u = uuid.New().String()
		n = "m-" + u + "." + ext
	)

	p := filepath.Join("images", n)

	of, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}

	defer of.Close()
	if _, err = of.Write(req.Data); err != nil {
		return "", err
	}

	return m.c.Root + p, nil
}
