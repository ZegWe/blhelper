package blhelper

import (
	"fmt"
	"log"
	"strconv"
)

// GetRoomInfo :
func (a *App) GetRoomInfo() error {
	req := a.newReq()
	req.Params = map[string]string{
		"mid": strconv.Itoa(a.Mid),
	}
	resp, err := req.Get("http://api.live.bilibili.com/room/v1/Room/getRoomInfoOld")
	if err != nil {
		return err
	}
	j, err := resp.Json()
	if err != nil {
		return err
	}
	id, err := j.Get("data").Get("roomid").Int()
	if err != nil {
		return err
	}
	liveStatus, err := j.Get("data").Get("liveStatus").Int()
	if err != nil {
		return err
	}
	title, err := j.Get("data").Get("title").String()
	if err != nil {
		return err
	}
	a.RoomID = id
	a.IsLiving = liveStatus == 1
	a.Title = title
	return nil
}

// StartLive :
func (a *App) StartLive(area string) (Rtmp, error) {
	req := a.newReq()
	req.Cookies = a.cookies
	resp, err := req.PostForm("http://api.live.bilibili.com/room/v1/Room/startLive", map[string]string{
		"room_id":  strconv.Itoa(a.RoomID),
		"area_v2":  area,
		"platform": "pc",
		"csrf":     a.csrf,
	})
	if err != nil {
		return Rtmp{}, err
	}
	j, err := resp.Json()
	if err != nil {
		return Rtmp{}, err
	}
	code, err := j.Get("code").Int()
	if err != nil {
		return Rtmp{}, err
	}
	if code != 0 {
		return Rtmp{}, fmt.Errorf("StartLive Error: %v", j)
	}
	log.Println("Live Start!")
	a.GetRoomInfo()
	return Rtmp{
		j.Get("data").Get("rtmp").Get("addr").MustString(),
		j.Get("data").Get("rtmp").Get("code").MustString(),
	}, nil
}

// StopLive :
func (a *App) StopLive() error {
	req := a.newReq()
	req.Cookies = a.cookies
	resp, err := req.PostForm("http://api.live.bilibili.com/room/v1/Room/stopLive", map[string]string{
		"room_id":  strconv.Itoa(a.RoomID),
		"platform": "pc",
		"csrf":     a.csrf,
	})
	if err != nil {
		return err
	}
	j, err := resp.Json()
	if err != nil {
		return err
	}
	code, err := j.Get("code").Int()
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("StopLive Error: %v", j)
	}
	log.Println("Live Stop!")
	a.GetRoomInfo()
	return nil
}

// SetLiveTitle returns title, error
func (a *App) SetLiveTitle(title string) (string, error) {
	req := a.newReq()
	req.Cookies = a.cookies
	resp, err := req.PostForm("http://api.live.bilibili.com/room/v1/Room/update", map[string]string{
		"room_id": strconv.Itoa(a.RoomID),
		"title":   title,
		"csrf":    a.csrf,
	})
	if err != nil {
		return "", err
	}
	j, err := resp.Json()
	if err != nil {
		return "", err
	}
	code, err := j.Get("code").Int()
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("SetLiveTitle Error: %v", j)
	}
	a.GetRoomInfo()
	log.Println("Live title set:", a.Title)
	return a.Title, nil
}
