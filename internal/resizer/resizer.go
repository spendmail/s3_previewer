package resizer

import (
	"bytes"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
)

type Resizer struct{}

// New is a resizer constructor.
func New() *Resizer {
	return &Resizer{}
}

const (
	MimePng  = "image/png"
	MimeJpeg = "image/jpeg"
)

func (r *Resizer) Resize(width, height uint, imageBytes []byte) ([]byte, error) {

	originalImage, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return []byte{}, err
	}

	newImage := resize.Resize(width, height, originalImage, resize.Lanczos3)
	buf := new(bytes.Buffer)

	mimeType := http.DetectContentType(imageBytes)
	if mimeType == MimePng {
		err = png.Encode(buf, newImage)
	} else if mimeType == MimeJpeg {
		err = jpeg.Encode(buf, newImage, nil)
	} else {
		err = errors.New("file type is not supported")
	}

	return buf.Bytes(), nil
}
