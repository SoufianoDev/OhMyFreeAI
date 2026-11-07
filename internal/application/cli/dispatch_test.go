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


func TestParseRecognizesBackupExportAndInspect(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Command
		path string
	}{
		{name: "export", args: []string{"backup", "export", "/tmp/state.omfaibackup"}, want: BackupExport, path: "/tmp/state.omfaibackup"},
		{name: "inspect", args: []string{"backup", "inspect", "/tmp/state.omfaibackup"}, want: BackupInspect, path: "/tmp/state.omfaibackup"},
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
			if len(invocation.Arguments) != 1 || invocation.Arguments[0] != test.path {
				t.Fatalf("arguments = %#v, want [%q]", invocation.Arguments, test.path)
			}
		})
	}
}

func TestParseRejectsMalformedBackupCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{name: "missing action", args: []string{"backup"}, want: ErrUnexpectedArguments},
		{name: "missing path", args: []string{"backup", "export"}, want: ErrUnexpectedArguments},
		{name: "extra path", args: []string{"backup", "inspect", "one.omfaibackup", "two.omfaibackup"}, want: ErrUnexpectedArguments},
		{name: "unknown action", args: []string{"backup", "restore", "state.omfaibackup"}, want: ErrUnknownCommand},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Parse(test.args); !errors.Is(err, test.want) {
				t.Fatalf("Parse(%v) error = %v, want %v", test.args, err, test.want)
			}
		})
	}
}
