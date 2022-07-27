package app

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"net/http"
	"testing"

	internalcache "github.com/spendmail/s3_previewer/internal/cache"
	internalconfig "github.com/spendmail/s3_previewer/internal/config"
	internallogger "github.com/spendmail/s3_previewer/internal/logger"
	internalresizer "github.com/spendmail/s3_previewer/internal/resizer"
	"github.com/stretchr/testify/require"
)

var (
	ImageWidth           = 300
	ImageHeight          = 200
	ImageURL             = "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/gopher_2000x1000.jpg"
	WrongImageURLPath    = "raw.githubusercontent.com/mistake_in_the_path"
	WrongDNSURL          = "this-is-non-existent-domain.com/image.jpeg"
	ContentTypeImageJpeg = "image/jpeg"
)

func TestApplication(t *testing.T) {
	t.Run("succeeding test", func(t *testing.T) {
		config, err := internalconfig.NewConfig("../../configs/previewer.toml")
		require.NoError(t, err, "should be without errors")

		logger, err := internallogger.New(config)
		require.NoError(t, err, "should be without errors")

		cache, err := internalcache.New(config, logger)
		require.NoError(t, err, "should be without errors")

		app, err := New(logger, internalresizer.New(), cache)
		require.NoError(t, err, "should be without errors")

		headers := map[string][]string{}
		imageBytes, err := app.ResizeImageByURL(ImageWidth, ImageHeight, ImageURL, headers)
		require.NoError(t, err, "should be without errors")

		bytesContentType := http.DetectContentType(imageBytes)
		require.Equal(t, ContentTypeImageJpeg, bytesContentType, fmt.Sprintf("content type should be %s, but %s given", ContentTypeImageJpeg, bytesContentType))

		img, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
		require.NoError(t, err, "should be without errors")
		require.Equal(t, ImageWidth, img.Width, fmt.Sprintf("image width should be %d, but %d given", ImageWidth, img.Width))
		require.Equal(t, ImageHeight, img.Height, fmt.Sprintf("image height should be %d, but %d given", ImageHeight, img.Height))
	})

	t.Run("wrong dns", func(t *testing.T) {
		config, err := internalconfig.NewConfig("../../configs/previewer.toml")
		require.NoError(t, err, "should be without errors")

		logger, err := internallogger.New(config)
		require.NoError(t, err, "should be without errors")

		cache, err := internalcache.New(config, logger)
		require.NoError(t, err, "should be without errors")

		app, err := New(logger, internalresizer.New(), cache)
		require.NoError(t, err, "should be without errors")

		headers := map[string][]string{}
		_, err = app.ResizeImageByURL(ImageWidth, ImageHeight, WrongDNSURL, headers)
		require.Truef(t, errors.Is(err, ErrServerNotExists), "actual error %q", err)
	})

	t.Run("file not found", func(t *testing.T) {
		config, err := internalconfig.NewConfig("../../configs/previewer.toml")
		require.NoError(t, err, "should be without errors")

		logger, err := internallogger.New(config)
		require.NoError(t, err, "should be without errors")

		cache, err := internalcache.New(config, logger)
		require.NoError(t, err, "should be without errors")

		app, err := New(logger, internalresizer.New(), cache)
		require.NoError(t, err, "should be without errors")

		headers := map[string][]string{}
		_, err = app.ResizeImageByURL(ImageWidth, ImageHeight, WrongImageURLPath, headers)
		require.Truef(t, errors.Is(err, ErrFileNotFound), "actual error %q", err)
	})
}
