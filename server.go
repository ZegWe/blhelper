package blhelper

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Default server config
const (
	DefaultServerPort = 2021
	DefaultServerIP   = "127.0.0.1"
)

// DefaultServerAddr : default server address
var DefaultServerAddr = fmt.Sprintf("%s:%d", DefaultServerIP, DefaultServerPort)

// Http Response Code
const (
	HTTPCodeSuccess int = iota
	HTTPCodeFail
)

// RunServer starts a http server for api controlling
func (a *App) RunServer() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/StopServer", httpStopServer(a))
	mux.HandleFunc("/LoginByQR", httpLogin(a))
	mux.HandleFunc("/CheckLogin", httpCheckLogin(a))
	mux.HandleFunc("/GetLiveStatus", httpGetLiveStatus(a))
	mux.HandleFunc("/SetTitle", httpSetTitle(a))
	mux.HandleFunc("/StartLive", httpStartLive(a))
	mux.HandleFunc("/StopLive", httpStopLive(a))

	a.server.Handler = mux
	return a.server.Serve(a.listener)
}

func httpStopServer(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")

		msg, _ := NewResponse(ResponseMsg{
			Code: HTTPCodeSuccess,
			Data: ResponseData{
				Message: "server stopped",
			},
		})
		rw.Write(msg)
		go func() {
			time.Sleep(time.Second)
			a.server.Shutdown(nil)
		}()
	}
}

func httpLogin(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")
		qr, url, err := a.LoginByQR()
		if err != nil {
			log.Println("http login error: ", err)
			msg, _ := NewResponse(ResponseMsg{
				Code:         HTTPCodeFail,
				ErrorMessage: err.Error(),
			})
			rw.Write(msg)
		} else {
			msg, _ := NewResponse(ResponseMsg{
				Code: HTTPCodeSuccess,
				Data: ResponseData{
					LoginQR: LoginQR{
						Msg: qr,
						URL: url,
					},
				},
			})
			rw.Write(msg)
		}
	}
}

func httpCheckLogin(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")
		msg, _ := NewResponse(ResponseMsg{
			Code: ExitCodeSuccess,
			Data: ResponseData{
				LoginStatus: a.IsLoggedIn,
			},
		})
		rw.Write(msg)
	}
}

func httpGetLiveStatus(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")
		a.GetRoomInfo()
		msg, _ := NewResponse(ResponseMsg{
			Code: HTTPCodeSuccess,
			Data: ResponseData{
				LiveStatus: LiveStatus{
					RoomID:   a.RoomID,
					Title:    a.Title,
					IsLiving: a.IsLiving,
				},
			},
		})
		rw.Write(msg)
	}
}

func httpSetTitle(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")
		title, err := a.SetLiveTitle(r.PostFormValue("title"))
		if err != nil {
			msg, _ := NewResponse(ResponseMsg{
				Code:         HTTPCodeFail,
				ErrorMessage: err.Error(),
			})
			rw.Write(msg)
		} else {
			msg, _ := NewResponse(ResponseMsg{
				Code: HTTPCodeSuccess,
				Data: ResponseData{
					RoomTitle: title,
				},
			})
			rw.Write(msg)
		}
	}
}

func httpStartLive(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")
		rtmp, err := a.StartLive(r.FormValue("area"))
		if err != nil {
			msg, _ := NewResponse(ResponseMsg{
				Code: HTTPCodeFail,
				ErrorMessage: err.Error(),
			})
			rw.Write(msg)
		}else {
			msg,_:=NewResponse(ResponseMsg{
				Code: HTTPCodeSuccess,
				Data: ResponseData{
					Rtmp: rtmp,
				},
			})
			rw.Write(msg)
		}
	}
}
func httpStopLive(a *App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/json")
		err := a.StopLive()
		if err != nil {
			msg,_:=NewResponse(ResponseMsg{
				Code: HTTPCodeFail,
				ErrorMessage: err.Error(),
			})
			rw.Write(msg)
		}else {
			msg,_:=NewResponse(ResponseMsg{
				Code: HTTPCodeSuccess,
				Data: ResponseData{
					Message: "live stopped",
				},
			})
			rw.Write(msg)
		}
	}
}
