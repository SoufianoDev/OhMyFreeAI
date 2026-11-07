package cli

import (
	"errors"
	"fmt"
	"strings"
)

type Command string

const (
	LaunchGUI     Command = "gui"
	ShowHelp      Command = "help"
	ShowVersion   Command = "version"
	RunDoctor     Command = "doctor"
	BackupExport  Command = "backup-export"
	BackupInspect Command = "backup-inspect"
)

var (
	ErrUnknownCommand       = errors.New("unknown command")
	ErrUnexpectedArguments = errors.New("unexpected command arguments")
)

type Invocation struct {
	Command   Command
	Arguments []string
}

// Parse keeps the desktop GUI as the default while making explicit CLI
// commands deterministic. Unknown commands are errors, never GUI fallbacks.
func Parse(args []string) (Invocation, error) {
	if len(args) == 0 {
		return Invocation{Command: LaunchGUI}, nil
	}

	if args[0] == "backup" {
		if len(args) != 3 || strings.TrimSpace(args[2]) == "" {
			return Invocation{}, fmt.Errorf("%w: expected backup export|inspect <archive.omfaibackup>", ErrUnexpectedArguments)
		}
		switch args[1] {
		case "export":
			return Invocation{Command: BackupExport, Arguments: []string{args[2]}}, nil
		case "inspect":
			return Invocation{Command: BackupInspect, Arguments: []string{args[2]}}, nil
		default:
			return Invocation{}, fmt.Errorf("%w: backup %q", ErrUnknownCommand, args[1])
		}
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
  OhMyFreeAI backup export <file.omfaibackup>  Export local database
  OhMyFreeAI backup inspect <file.omfaibackup> Validate a backup archive

Unknown commands are rejected and never launch the desktop interface.
`
