package blhelper

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// LoginByQR : returns qr, url, error
func (a *App) LoginByQR() (string, string, error) {
	url, key, err := a.getLoginURLAndKey()
	if err != nil {
		return "", "", err
	}
	qr, err := NewQR(url)
	if err != nil {
		return "", "", err
	}
	log.Println(url)
	fmt.Print(qr.OutPut())
	go func() {
		for i := 0; i < 30; i++ {
			log.Printf("Scan QR code to log in")
			err := a.tryLogin(key)
			if err == nil {
				log.Println("Login succeed!")
				a.IsLoggedIn = true
				break
			}
			log.Println(err)
			time.Sleep(time.Second)
		}
	}()
	return qr.OutPut(), url, nil
}

// WaitLogin :
func (a *App) WaitLogin() error {
	for i := 0; i < 30; i++ {
		if a.IsLoggedIn {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errLoginTimeOut
}

func (a *App) getLoginURLAndKey() (string, string, error) {
	req := a.newReq()
	resp, err := req.Get("http://passport.bilibili.com/qrcode/getLoginUrl")
	if err != nil {
		return "", "", err
	}
	j, err := resp.Json()
	if err != nil {
		return "", "", err
	}
	url, err := j.Get("data").Get("url").String()
	if err != nil {
		return "", "", err
	}
	key, err := j.Get("data").Get("oauthKey").String()
	if err != nil {
		return "", "", err
	}
	return url, key, nil
}

func (a *App) tryLogin(key string) error {
	req := a.newReq()
	// req.Data = map[string]string{
	// 	"oauthKey": key,
	// }
	resp, err := req.PostForm("http://passport.bilibili.com/qrcode/getLoginInfo", map[string]string{
		"oauthKey": key,
	})
	if err != nil {
		return err
	}
	cookies := resp.Cookies()
	id, sess, jct := "", "", ""
	for k := range a.cookies {
		delete(a.cookies, k)
	}
	for i := 0; i < len(cookies); i++ {
		if cookies[i].Name == "bili_jct" {
			jct = cookies[i].Value
		}
		if cookies[i].Name == "SESSDATA" {
			sess = cookies[i].Value
		}
		if cookies[i].Name == "DedeUserID" {
			id = cookies[i].Value
		}
		a.cookies[cookies[i].Name] = cookies[i].Value
		log.Println(cookies[i].Name, ":\t", cookies[i].Value)
	}
	if id == "" || sess == "" || jct == "" {
		return errNoCookieFound
	}
	a.Mid, err = strconv.Atoi(id)
	if err != nil {
		return err
	}
	a.csrf = jct
	return nil
}

func createQRFile(path string, url string) error {
	buf, err := qrcode.Encode(url, qrcode.High, 256)

	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(buf)
	return nil
}
