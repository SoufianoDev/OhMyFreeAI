package cli

import (
	"errors"
	"testing"
)

func TestParseDefaultsToGUIWithoutArguments(t *testing.T) {
	invocation, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse(nil) returned error: %v", err)
	}
	if invocation.Command != LaunchGUI {
		t.Fatalf("command = %q, want %q", invocation.Command, LaunchGUI)
	}
}

func TestParseRecognizesSupportedNonGUICommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Command
	}{
		{name: "help flag", args: []string{"--help"}, want: ShowHelp},
		{name: "short help flag", args: []string{"-h"}, want: ShowHelp},
		{name: "version flag", args: []string{"--version"}, want: ShowVersion},
		{name: "short version flag", args: []string{"-v"}, want: ShowVersion},
		{name: "doctor command", args: []string{"doctor"}, want: RunDoctor},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invocation, err := Parse(test.args)
			if err != nil {
				t.Fatalf("Parse(%v) returned error: %v", test.args, err)
			}
			if invocation.Command != test.want {
				t.Fatalf("command = %q, want %q", invocation.Command, test.want)
			}
		})
	}
}

func TestParseRejectsUnknownCommandsWithoutFallingBackToGUI(t *testing.T) {
	invocation, err := Parse([]string{"runtime"})
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("error = %v, want ErrUnknownCommand", err)
	}
	if invocation.Command == LaunchGUI {
		t.Fatal("unknown command must never dispatch the GUI")
	}
}

func TestParseRejectsUnexpectedArguments(t *testing.T) {
	_, err := Parse([]string{"doctor", "--verbose"})
	if !errors.Is(err, ErrUnexpectedArguments) {
		t.Fatalf("error = %v, want ErrUnexpectedArguments", err)
	}
}
