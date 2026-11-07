package cli

import (
	"errors"
	"fmt"
)

type Command string

const (
	LaunchGUI  Command = "gui"
	ShowHelp   Command = "help"
	ShowVersion Command = "version"
	RunDoctor  Command = "doctor"
)

var (
	ErrUnknownCommand       = errors.New("unknown command")
	ErrUnexpectedArguments = errors.New("unexpected command arguments")
)

type Invocation struct {
	Command Command
}

// Parse keeps the desktop GUI as the default while making explicit CLI
// commands deterministic. Unknown commands are errors, never GUI fallbacks.
func Parse(args []string) (Invocation, error) {
	if len(args) == 0 {
		return Invocation{Command: LaunchGUI}, nil
	}
	if len(args) > 1 {
		return Invocation{}, fmt.Errorf("%w: %q", ErrUnexpectedArguments, args[1:])
	}

	switch args[0] {
	case "--help", "-h":
		return Invocation{Command: ShowHelp}, nil
	case "--version", "-v":
		return Invocation{Command: ShowVersion}, nil
	case "doctor":
		return Invocation{Command: RunDoctor}, nil
	default:
		return Invocation{}, fmt.Errorf("%w: %q", ErrUnknownCommand, args[0])
	}
}

const Usage = `OhMyFreeAI

Usage:
  OhMyFreeAI              Launch the desktop application
  OhMyFreeAI --help       Show this help
  OhMyFreeAI --version    Show application version
  OhMyFreeAI doctor       Inspect local application environment

Unknown commands are rejected and never launch the desktop interface.
`
