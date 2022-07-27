package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var (
	HTTPHost             = os.Getenv("TESTS_HTTP_HOST")
	ImageWidth           = 300
	ImageHeight          = 200
	HTTPHostPath         = fmt.Sprintf("/fill/%d/%d/", ImageWidth, ImageHeight)
	ImageURL             = "raw.githubusercontent.com/OtusGolang/final_project/master/examples/image-previewer/gopher_2000x1000.jpg"
	WrongImageURLPath    = "raw.githubusercontent.com/mistake_in_the_path"
	WrongDNSURL          = "this-is-non-existent-domain.com/image.jpeg"
	ContentTypeImageJpeg = "image/jpeg"
	ContentTypeTextPlain = "text/plain"
)

func init() {
	if HTTPHost == "" {
		HTTPHost = "http://localhost:8888"
	}
}

func TestHTTP(t *testing.T) {
	t.Run("http status 200", func(t *testing.T) {
		u, err := url.Parse(HTTPHost)
		require.NoError(t, err, "should be without errors")

		u.Path = path.Join(u.Path, HTTPHostPath, ImageURL)
		testingURL := u.String()

		t.Logf("Requesting %v\n", testingURL)
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testingURL, nil)
		require.NoError(t, err, "should be without errors")

		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err, "should be without errors")
		defer response.Body.Close()

		require.Equal(t, response.StatusCode, http.StatusOK, fmt.Sprintf("response status code should be %d, but %d given", http.StatusOK, response.StatusCode))
		httpContentType := response.Header.Get("Content-Type")
		require.Equal(t, ContentTypeImageJpeg, httpContentType, fmt.Sprintf("content type should be %s, but %s given", ContentTypeImageJpeg, httpContentType))

		responseBytes, err := io.ReadAll(response.Body)
		bytesContentType := http.DetectContentType(responseBytes)
		require.NoError(t, err, "should be without errors")
		require.Equal(t, ContentTypeImageJpeg, bytesContentType, fmt.Sprintf("content type should be %s, but %s given", ContentTypeImageJpeg, bytesContentType))

		img, _, err := image.DecodeConfig(bytes.NewReader(responseBytes))
		require.NoError(t, err, "should be without errors")
		require.Equal(t, ImageWidth, img.Width, fmt.Sprintf("image width should be %d, but %d given", ImageWidth, img.Width))
		require.Equal(t, ImageHeight, img.Height, fmt.Sprintf("image height should be %d, but %d given", ImageHeight, img.Height))
	})

	t.Run("wrong image url", func(t *testing.T) {
		u, err := url.Parse(HTTPHost)
		require.NoError(t, err, "should be without errors")

		u.Path = path.Join(u.Path, HTTPHostPath, WrongImageURLPath)
		testingURL := u.String()

		t.Logf("Requesting %v\n", testingURL)
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testingURL, nil)
		require.NoError(t, err, "should be without errors")

		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err, "should be without errors")
		defer response.Body.Close()

		require.Equal(t, response.StatusCode, http.StatusBadGateway, fmt.Sprintf("response status code should be %d, but %d given", http.StatusBadGateway, response.StatusCode))
		httpContentType := response.Header.Get("Content-Type")
		require.True(t, strings.HasPrefix(httpContentType, ContentTypeTextPlain), fmt.Sprintf("content type should be %s, but %s given", ContentTypeTextPlain, httpContentType))

		responseBytes, err := io.ReadAll(response.Body)
		require.NoError(t, err, "should be without errors")
		bytesContentType := http.DetectContentType(responseBytes)
		require.True(t, strings.HasPrefix(bytesContentType, ContentTypeTextPlain), fmt.Sprintf("content type should be %s, but %s given", ContentTypeTextPlain, bytesContentType))
	})

	t.Run("wrong dns", func(t *testing.T) {
		u, err := url.Parse(HTTPHost)
		require.NoError(t, err, "should be without errors")

		u.Path = path.Join(u.Path, HTTPHostPath, WrongDNSURL)
		testingURL := u.String()

		t.Logf("Requesting %v\n", testingURL)
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testingURL, nil)
		require.NoError(t, err, "should be without errors")

		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err, "should be without errors")
		defer response.Body.Close()

		require.Equal(t, response.StatusCode, http.StatusBadGateway, fmt.Sprintf("response status code should be %d, but %d given", http.StatusBadGateway, response.StatusCode))
		httpContentType := response.Header.Get("Content-Type")
		require.True(t, strings.HasPrefix(httpContentType, ContentTypeTextPlain), fmt.Sprintf("content type should be %s, but %s given", ContentTypeTextPlain, httpContentType))

		responseBytes, err := io.ReadAll(response.Body)
		require.NoError(t, err, "should be without errors")
		bytesContentType := http.DetectContentType(responseBytes)
		require.True(t, strings.HasPrefix(bytesContentType, ContentTypeTextPlain), fmt.Sprintf("content type should be %s, but %s given", ContentTypeTextPlain, bytesContentType))
	})
}
