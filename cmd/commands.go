package cmd

import (
	"flag"
	"regexp"
)

// Command :
type Command struct {
	Name  string
	Func  CommandFunc
	Short string
	Long  string
	Usage string
	Flags *flag.FlagSet
}

var commands = make(map[string]Command)

func init() {
	RegisterCommand(Command{
		Name:  "help",
		Func:  cmdHelp,
		Short: "show command help",
		Usage: "<command>",
	})
	RegisterCommand(Command{
		Name: "start",
		Func: cmdStart,
		Short: "start in background",
		Usage: "",
	})
	RegisterCommand(Command{
		Name: "run",
		Func: cmdRun,
		Short: "run server forground",
		Usage: "[--pingback <address>]",
		Flags: func () *flag.FlagSet {
			fl := flag.NewFlagSet("run", flag.ExitOnError)
			fl.String("pingback", "", "ping back to given address when start")
			return fl
		}(),
	})
	RegisterCommand(Command{
		Name: "stop",
		Func: cmdStop,
		Short: "stop server",
		Usage: "",
	})
	RegisterCommand(Command{
		Name: "login",
		Func: cmdLogin,
		Short: "login by qr",
		Usage: "",
	})
	RegisterCommand(Command{
		Name: "live",
		Func: cmdLive,
		Short: "manage live room",
		Usage: "[--info] [--set-title <title>] [--start <area_number>] [--stop]",
		Flags: func () *flag.FlagSet {
			fl := flag.NewFlagSet("live", flag.ExitOnError)
			fl.Bool("info", false, "get live room info")
			fl.String("set-title", "", "update live title")
			fl.String("start", "", "start live in given area")
			fl.Bool("stop", false, "stop live")
			return fl
		}(),
	})
}

// RegisterCommand registers the command cmd.
// cmd.Name must be unique and conform to the
// following format:
//
//    - lowercase
//    - alphanumeric and hyphen characters only
//    - cannot start or end with a hyphen
//    - hyphen cannot be adjacent to another hyphen
//
// This function panics if the name is already registered,
// if the name does not meet the described format, or if
// any of the fields are missing from cmd.
//
// This function should be used in init().
func RegisterCommand(cmd Command) {
	if cmd.Name == "" {
		panic("command name is required")
	}
	if cmd.Func == nil {
		panic("command function missing")
	}
	if cmd.Short == "" {
		panic("command short string is required")
	}
	if _, exists := commands[cmd.Name]; exists {
		panic("command already registered: " + cmd.Name)
	}
	if !commandNameRegex.MatchString(cmd.Name) {
		panic("invalid command name")
	}
	commands[cmd.Name] = cmd
}

var commandNameRegex = regexp.MustCompile(`^[a-z0-9]$|^([a-z0-9]+-?[a-z0-9]*)+[a-z0-9]$`)
