package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	DefaultScheme = "http://"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Resizer interface {
	Resize(width, height uint, image []byte) ([]byte, error)
}

type Cache interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Clear()
}

type S3Client interface {
	Download(context context.Context, bucket, key string) ([]byte, error)
}

type Application struct {
	Logger   Logger
	Resizer  Resizer
	Cache    Cache
	S3Client S3Client
}

var (
	ErrDownload        = errors.New("unable to download a file")
	ErrFileNotFound    = errors.New("file not found")
	ErrServerNotExists = errors.New("remove server doesn't exist")
	ErrRequest         = errors.New("request error")
	ErrFileRead        = errors.New("unable to read a file")
)

// New is an application constructor.
func New(logger Logger, resizer Resizer, cache Cache, s3Client S3Client) (*Application, error) {
	return &Application{
		Cache:    cache,
		Logger:   logger,
		Resizer:  resizer,
		S3Client: s3Client,
	}, nil
}

func goId() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// ResizeImageByURL downloads, caches and crops images by given sizes and URL.
func (app *Application) ResizeImageByURL(width, height int, bucket string, key string, headers map[string][]string) ([]byte, error) {
	// Key includes sizes in order to store different files for different sizes of the same file.
	//hash := md5.Sum([]byte(fmt.Sprintf("%s-%s-%d-%d", bucket, key, width, height)))
	//cacheKey := hex.EncodeToString(hash[:])

	//// If file exists in cache, return from there.
	//resultBytes, err := app.Cache.Get(cacheKey)
	//if err == nil {
	//	return resultBytes, nil
	//}

	sourceBytes, err := app.S3Client.Download(context.TODO(), bucket, key)
	if err != nil {
		return []byte{}, err
	}

	originalImage, _, err := image.Decode(bytes.NewReader(sourceBytes))
	if err != nil {
		return []byte{}, err
	}

	newImage := resize.Resize(uint(width), uint(height), originalImage, resize.Lanczos3)
	buf := new(bytes.Buffer)

	ext := strings.ToLower(filepath.Ext(key))
	if ext == ".png" {
		err = png.Encode(buf, newImage)
	} else if ext == ".jpg" || ext == ".jpeg" {
		err = jpeg.Encode(buf, newImage, nil)
	} else {
		err = errors.New(fmt.Sprintf("file type %s is not supported", key))
	}

	if err != nil {
		return []byte{}, err
	}

	// Set processed image in cache
	//_ = app.Cache.Set(cacheKey, resultBytes)

	// And return slice of bytes.
	return buf.Bytes(), nil
}

// downloadByURL downloads image by given url forwarding original headers.
func (app *Application) downloadByURL(url string, headers map[string][]string) ([]byte, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, DefaultScheme+url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrRequest, err)
	}

	// Forwarding original headers to remote server.
	for name, values := range headers {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		// Identifying wrong domain name errors.
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			return []byte{}, fmt.Errorf("%w: %s", ErrServerNotExists, err)
		}

		return []byte{}, fmt.Errorf("%w: %s", ErrDownload, err)
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrFileRead, err)
	}

	return bytes, nil
}
