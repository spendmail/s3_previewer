package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	URLResizePattern = "/resize/{width:[0-9]+}/{height:[0-9]+}/{bucket:[a-zA-Z-]+}/{key:.+}"
	WidthField       = "width"
	HeightField      = "height"
	BucketField      = "bucket"
	KeyField         = "key"
)

type Config interface {
	GetHTTPHost() string
	GetHTTPPort() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Application interface {
	ResizeImageByURL(width, height int, bucket string, key string, headers map[string][]string) ([]byte, error)
}

type Server struct {
	Logger Logger
	Server *http.Server
}

var (
	ErrParameterParseWidth  = errors.New("unable to parse image width")
	ErrParameterParseHeight = errors.New("unable to parse image height")
	ErrResizeImage          = errors.New("unable to resize an image")
	ErrResponseWrite        = errors.New("unable to write a response")
)

type Handler struct {
	App    Application
	Logger Logger
}

// New is HTTP service constructor.
func New(config Config, logger Logger, app Application) *Server {
	handler := &Handler{
		App:    app,
		Logger: logger,
	}

	router := mux.NewRouter()
	router.HandleFunc(URLResizePattern, handler.resizeHandler).Methods(http.MethodGet)

	server := &http.Server{
		Addr:    net.JoinHostPort(config.GetHTTPHost(), config.GetHTTPPort()),
		Handler: router,
	}

	return &Server{
		Logger: logger,
		Server: server,
	}
}

// resizeHandler handles cropping requests.
func (h *Handler) resizeHandler(w http.ResponseWriter, r *http.Request) {
	width, err := strconv.Atoi(mux.Vars(r)[WidthField])
	if err != nil {
		SendBadGatewayStatus(w, h, fmt.Errorf("%w: %s", ErrParameterParseWidth, err))
		return
	}

	height, err := strconv.Atoi(mux.Vars(r)[HeightField])
	if err != nil {
		SendBadGatewayStatus(w, h, fmt.Errorf("%w: %s", ErrParameterParseHeight, err))
		return
	}

	bytes, err := h.App.ResizeImageByURL(width, height, mux.Vars(r)[BucketField], mux.Vars(r)[KeyField], r.Header)
	if err != nil {
		SendBadGatewayStatus(w, h, err)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(bytes))
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	if _, err := w.Write(bytes); err != nil {
		h.Logger.Error(fmt.Errorf("%w: %s", ErrResizeImage, err.Error()))
	}
}

// SendBadGatewayStatus sends http.StatusBadGateway response with custom message.
func SendBadGatewayStatus(w http.ResponseWriter, h *Handler, err error) {
	w.WriteHeader(http.StatusBadGateway)
	if n, e := w.Write([]byte(err.Error())); e != nil {
		h.Logger.Error(fmt.Errorf("%w: trying to write %d bytes: %s", ErrResponseWrite, n, e.Error()))
	}
	h.Logger.Error(err.Error())
}

// Start launches a HTTP server.
func (s *Server) Start() error {
	return s.Server.ListenAndServe()
}

// Stop suspends HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
