package resizer

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	ImageWidth  = 300
	ImageHeight = 200
	Scheme      = "http://"
	ImageURL    = "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/gopher_2000x1000.jpg"
)

func TestResizer(t *testing.T) {
	t.Run("resizing test", func(t *testing.T) {
		resizer := New()

		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, Scheme+ImageURL, nil)
		require.NoError(t, err, "should be without errors")

		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err, "should be without errors")
		defer response.Body.Close()

		imageBytes, err := io.ReadAll(response.Body)
		require.NoError(t, err, "should be without errors")

		croppedImageBytes, err := resizer.Resize(uint(ImageWidth), uint(ImageHeight), imageBytes)
		require.NoError(t, err, "should be without errors")

		img, _, err := image.DecodeConfig(bytes.NewReader(croppedImageBytes))
		require.NoError(t, err, "should be without errors")
		require.Equal(t, ImageWidth, img.Width, fmt.Sprintf("image width should be %d, but %d given", ImageWidth, img.Width))
		require.Equal(t, ImageHeight, img.Height, fmt.Sprintf("image height should be %d, but %d given", ImageHeight, img.Height))
	})
}
