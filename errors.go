package blhelper

import "errors"

var (
	errNoCookieFound = errors.New("no coockie found")
	errGetRoomID     = errors.New("failed to get room id")
	errLoginTimeOut  = errors.New("login time out")
)

// Exit Code Enum
const (
	ExitCodeSuccess int = iota
	ExitCodeErrorStart
	ExitCodeErrorRun
)
