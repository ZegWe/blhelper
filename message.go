package blhelper

import "encoding/json"

// ResponseMsg : http response
type ResponseMsg struct {
	Code         int          `json:"code"`
	ErrorMessage string       `json:"error_msg"`
	Data         ResponseData `json:"data"`
}

// ResponseData :
type ResponseData struct {
	LoginStatus bool       `json:"login_status"`
	LoginQR     LoginQR    `json:"login_qr"`
	Message     string     `json:"msg"`
	LiveStatus  LiveStatus `json:"live_status"`
	RoomTitle   string     `json:"room_title"`
	Rtmp        Rtmp       `json:"rtmp"`
}

// LoginQR :
type LoginQR struct {
	Msg string `json:"msg"`
	URL string `json:"url"`
}

// LiveStatus :
type LiveStatus struct {
	RoomID   int    `json:"room_id"`
	Title    string `json:"tile"`
	IsLiving bool   `json:"is_living"`
}

// Rtmp :
type Rtmp struct {
	Addr string `json:"addr"`
	Code string `json:"code"`
}

// NewResponse :
func NewResponse(msg ResponseMsg) ([]byte, error) {
	return json.Marshal(msg)
}
