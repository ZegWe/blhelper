package blhelper

import (
	"net"
	"net/http"

	"github.com/mozillazg/request"
)

// App : Bilibili Live Helper
type App struct {
	listener   net.Listener
	server     *http.Server
	client     *http.Client
	cookies    map[string]string
	csrf       string
	Mid        int
	RoomID     int
	Title      string
	IsLiving   bool
	IsLoggedIn bool
}

// NewApp : initialized new app
func NewApp() (*App, error) {
	ln, err := net.Listen("tcp", DefaultServerAddr)
	if err != nil {
		return nil, err
	}
	return &App{
		listener:   ln,
		server:     &http.Server{},
		client:     &http.Client{},
		cookies:    make(map[string]string),
		csrf:       "",
		Mid:        0,
		RoomID:     0,
		Title:      "",
		IsLiving:   false,
		IsLoggedIn: false,
	}, nil
}

func (a App) newReq() *request.Request {
	return request.NewRequest(a.client)
}
