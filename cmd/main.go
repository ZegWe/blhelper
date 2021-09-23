package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/zegwe/blhelper"
)

// Main entrance
func Main() {
	switch len(os.Args) {
	case 0:
		fmt.Printf("[FATAL] no arguments provided by OS; args[0] must be command\n")
		os.Exit(blhelper.ExitCodeErrorStart)
	case 1:
		os.Args = append(os.Args, "help")
	}

	cmdName := os.Args[1]
	subCommand, ok := commands[cmdName]
	if !ok {
		fmt.Printf("[ERROR] '%s' is not a recognized subcommand; see 'blhelper help'\n", os.Args[1])
		os.Exit(blhelper.ExitCodeErrorStart)
	}
	flags := subCommand.Flags
	if flags == nil {
		flags = flag.NewFlagSet(subCommand.Name, flag.ExitOnError)
	}

	err := flags.Parse(os.Args[2:])
	if err != nil {
		fmt.Println(err)
		os.Exit(blhelper.ExitCodeErrorStart)
	}

	exitCode, err := subCommand.Func(Flags{flags})
	if err != nil {
		log.Printf("%v: %v", subCommand.Name, err)
	}

	os.Exit(exitCode)
}

// Flags wraps a FlagSet so that typed values
// from flags can be easily retrieved.
type Flags struct {
	*flag.FlagSet
}

// String returns the string representation of the
// flag given by name. It panics if the flag is not
// in the flag set.
func (f Flags) String(name string) string {
	return f.FlagSet.Lookup(name).Value.String()
}

// Bool returns the boolean representation of the
// flag given by name. It returns false if the flag
// is not a boolean type. It panics if the flag is
// not in the flag set.
func (f Flags) Bool(name string) bool {
	val, _ := strconv.ParseBool(f.String(name))
	return val
}

// Int returns the integer representation of the
// flag given by name. It returns 0 if the flag
// is not an integer type. It panics if the flag is
// not in the flag set.
func (f Flags) Int(name string) int {
	val, _ := strconv.ParseInt(f.String(name), 0, strconv.IntSize)
	return int(val)
}

// Float64 returns the float64 representation of the
// flag given by name. It returns false if the flag
// is not a float64 type. It panics if the flag is
// not in the flag set.
func (f Flags) Float64(name string) float64 {
	val, _ := strconv.ParseFloat(f.String(name), 64)
	return val
}

// flagHelp returns the help text for fs.
func flagHelp(fs *flag.FlagSet) string {
	if fs == nil {
		return ""
	}

	// temporarily redirect output
	out := fs.Output()
	defer fs.SetOutput(out)

	buf := new(bytes.Buffer)
	fs.SetOutput(buf)
	fs.PrintDefaults()
	return buf.String()
}
