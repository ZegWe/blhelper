package cmd

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/mozillazg/request"
	"github.com/zegwe/blhelper"
)

// CommandFunc : command line function
type CommandFunc func(Flags) (int, error)

func cmdStart(fl Flags) (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return blhelper.ExitCodeErrorRun,
			fmt.Errorf("open ping back listener err: %v", err)
	}
	defer ln.Close()
	fmt.Println(ln.Addr().String())
	cmd := exec.Command(os.Args[0], "run", "--pingback", ln.Addr().String())

	stdinpipe, err := cmd.StdinPipe()
	if err != nil {
		return blhelper.ExitCodeErrorStart,
			fmt.Errorf("create stdin pipe err: %v", err)
	}

	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	pingBackBytes := make([]byte, 32)
	_, err = rand.Read(pingBackBytes)
	if err != nil {
		return blhelper.ExitCodeErrorStart,
			fmt.Errorf("generate ping back bytes err: %v", err)
	}
	log.Printf("generate bytes: %v\n", pingBackBytes)
	go func() {
		stdinpipe.Write(pingBackBytes)
		stdinpipe.Close()
	}()

	err = cmd.Start()
	if err != nil {
		return blhelper.ExitCodeErrorStart,
			fmt.Errorf("start blhelper process err: %v", err)
	}

	success, exit := make(chan struct{}), make(chan error)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Println(err)
				}
				break
			}

			err = handlePingbackConn(conn, pingBackBytes)
			if err == nil {
				close(success)
				break
			}
			log.Println(err)
		}
	}()

	go func() {
		err := cmd.Wait()
		exit <- err
	}()

	select {
	case <-success:
		fmt.Printf("Successfully started blhelper (pid=%d) in background\n", cmd.Process.Pid)
	case err := <-exit:
		return blhelper.ExitCodeErrorStart,
			fmt.Errorf("blhelper process exited with err: %v", err)
	}
	return blhelper.ExitCodeSuccess, nil
}

func cmdRun(fl Flags) (int, error) {
	pingBackFlag := fl.String("pingback")

	app, err := blhelper.NewApp()
	if err != nil {
		return blhelper.ExitCodeErrorStart, err
	}
	if pingBackFlag != "" {
		pingBackBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return blhelper.ExitCodeErrorStart,
				fmt.Errorf("reading ping back bytes err: %v", err)
		}
		conn, err := net.Dial("tcp", pingBackFlag)
		if err != nil {
			return blhelper.ExitCodeErrorStart,
				fmt.Errorf("dialing ping back address(%v) err: %v", pingBackFlag, err)
		}
		defer conn.Close()
		_, err = conn.Write(pingBackBytes)
		if err != nil {
			return blhelper.ExitCodeErrorStart,
				fmt.Errorf("ping back to %v err: %v", pingBackFlag, err)
		}
	}

	err = app.RunServer()
	if err != nil {
		return blhelper.ExitCodeErrorRun, err
	}

	return blhelper.ExitCodeSuccess, nil
}

func cmdLive(fl Flags) (int, error) {
	startFlag := fl.String("start")
	stopFlag := fl.Bool("stop")
	titleFlag := fl.String("set-title")
	infoFlag := fl.Bool("info")
	if startFlag != "" {
		resp, err := apiRequest("", http.MethodPost, "/StartLive", map[string]string{"area": startFlag})
		if err != nil {
			return blhelper.ExitCodeErrorRun, err
		}
		if resp.Code != blhelper.HTTPCodeSuccess {
			return blhelper.ExitCodeErrorRun, fmt.Errorf("start live err: %v", resp.ErrorMessage)
		}
		fmt.Printf("live started\nrtmp address:\t%s\ncode:\t%s\n", resp.Data.Rtmp.Addr, resp.Data.Rtmp.Code)
		return blhelper.ExitCodeSuccess, nil
	}

	if stopFlag != false {
		resp, err := apiRequest("", http.MethodGet, "/StopLive", nil)
		if err != nil {
			return blhelper.ExitCodeErrorRun, err
		}
		if resp.Code != blhelper.HTTPCodeSuccess {
			return blhelper.ExitCodeErrorRun, fmt.Errorf("stop live err: %v", resp.ErrorMessage)
		}
		fmt.Println(resp.Data.Message)
		return blhelper.ExitCodeSuccess, nil
	}

	if titleFlag != "" {
		resp, err := apiRequest("", http.MethodPost, "/SetTitle", map[string]string{"title": titleFlag})
		if err != nil {
			return blhelper.ExitCodeErrorRun, err
		}
		if resp.Code != blhelper.HTTPCodeSuccess {
			return blhelper.ExitCodeErrorRun, fmt.Errorf("set title err: %v", resp.ErrorMessage)
		}
		fmt.Printf("title updated: %v\n", resp.Data.RoomTitle)
		return blhelper.ExitCodeSuccess, nil
	}

	if infoFlag != false {
		resp, err := apiRequest("", http.MethodGet, "/GetLiveStatus", nil)
		if err != nil {
			return blhelper.ExitCodeErrorRun, err
		}
		if resp.Code != blhelper.HTTPCodeSuccess {
			return blhelper.ExitCodeErrorRun, fmt.Errorf("get info err: %v", resp.ErrorMessage)
		}
		fmt.Printf("Room ID: %v\nRoom Title: %v\nIsLiving: %v\n", resp.Data.LiveStatus.RoomID, resp.Data.LiveStatus.Title, resp.Data.LiveStatus.IsLiving)
		return blhelper.ExitCodeSuccess, nil

	}

	return blhelper.ExitCodeErrorRun, fmt.Errorf("use 'blhelper help live' for more information")
}

func cmdStop(fl Flags) (int, error) {
	resp, err := apiRequest("", http.MethodGet, "/StopServer", nil)
	if err != nil {
		return blhelper.ExitCodeErrorRun, err
	}
	if resp.Code != blhelper.HTTPCodeSuccess {
		return blhelper.ExitCodeErrorRun, fmt.Errorf("stop server failed: %s", resp.ErrorMessage)
	}
	fmt.Println(resp.Data.Message)
	return blhelper.ExitCodeSuccess, nil
}

func cmdLogin(fl Flags) (int, error) {
	resp, err := apiRequest("", http.MethodGet, "/LoginByQR", nil)
	if err != nil {
		return blhelper.ExitCodeErrorRun, err
	}
	if resp.Code != blhelper.HTTPCodeSuccess {
		return blhelper.ExitCodeErrorRun, fmt.Errorf("login err: %s", resp.ErrorMessage)
	}
	fmt.Printf("login url: %s\n%v\n", resp.Data.LoginQR.URL, resp.Data.LoginQR.Msg)
	for i := 0; i < 30; i++ {
		resp, err := apiRequest("", http.MethodGet, "/CheckLogin", nil)
		if err != nil {
			return blhelper.ExitCodeErrorRun, err
		}
		if resp.Code != blhelper.HTTPCodeSuccess {
			return blhelper.ExitCodeErrorRun, fmt.Errorf("CheckLogin err: %s", resp.ErrorMessage)
		}
		if resp.Data.LoginStatus {
			fmt.Println("login succeed!")
			return blhelper.ExitCodeSuccess, nil
		}
		time.Sleep(time.Second)
	}
	return blhelper.ExitCodeErrorRun, fmt.Errorf("login time out")
}

func cmdHelp(fl Flags) (int, error) {
	args := fl.Args()
	if len(args) == 0 {
		s := `blhelper(Bilibili Live Helper) is a simple tool which allows you to manage your live room easily.

Usage:
    blhelper <command> [<args...>]

Commands:
`
		keys := make([]string, 0, len(commands))
		for k := range commands {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			cmd := commands[k]
			s += fmt.Sprintf("    %-15s  %s\n", cmd.Name, cmd.Short)
		}

		s += "\nUse 'blhelper help <command>' for more information\n"

		fmt.Print(s)

		return blhelper.ExitCodeSuccess, nil
	} else if len(args) > 1 {
		return blhelper.ExitCodeErrorStart, fmt.Errorf("Too many commands given")
	}

	subCommand, ok := commands[args[0]]
	if !ok {
		return blhelper.ExitCodeErrorStart, fmt.Errorf("No such command: %s", args[0])
	}

	helpText := subCommand.Long
	if helpText == "" {
		helpText = subCommand.Short
	}

	result := fmt.Sprintf("%s\n\nUsage:\n    blhelper %s %s\n",
		helpText,
		subCommand.Name,
		subCommand.Usage,
	)

	if help := flagHelp(subCommand.Flags); help != "" {
		result += fmt.Sprintf("\nFlags:\n%s\n", help)
	}
	fmt.Print(result)

	return blhelper.ExitCodeSuccess, nil
}

func apiRequest(addr string, method string, url string, data map[string]string) (*blhelper.ResponseMsg, error) {
	req := request.NewRequest(http.DefaultClient)
	if addr == "" {
		addr = blhelper.DefaultServerAddr
	}
	reqAddr := "http://" + addr + url
	var resp *request.Response
	var err error
	switch method {
	case http.MethodPost:
		resp, err = req.PostForm(reqAddr, data)
	default:
		resp, err = req.Get(reqAddr)
	}
	if err != nil {
		return &blhelper.ResponseMsg{}, fmt.Errorf("api request err: %v", err)
	}
	byt, _ := resp.Content()
	var respObj blhelper.ResponseMsg
	json.Unmarshal(byt, &respObj)
	return &respObj, nil
}

// handlePingbackConn reads from conn and ensures it matches
// the bytes in expect, or returns an error if it doesn't.
func handlePingbackConn(conn net.Conn, expect []byte) error {
	defer conn.Close()
	confirmationBytes, err := ioutil.ReadAll(io.LimitReader(conn, 32))
	if err != nil {
		return err
	}
	if !bytes.Equal(confirmationBytes, expect) {
		return fmt.Errorf("wrong confirmation: %x", confirmationBytes)
	}
	return nil
}
