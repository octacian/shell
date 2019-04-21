package shell

import (
	"strings"
	"testing"
)

// output is used as a dummy output for printing during tests.
var output = &strings.Builder{}

// app is used throughout tests to ensure proper functionality.
var app = &App{Name: "command_test", Output: output, ErrOutput: output}

// TmplCmdWithSubCmd is used throughout tests to ensure proper functionality.
var TmplCmdWithSubCmd = Command{
	Name: "test",
	Usage: `${name} ${shortFlags}:

Execute primary test command.

${flags}`,
	SetFlags: func(ctx *Context) {
		ctx.Set("top", ctx.FlagSet().Int("top", 12, "example top-level flag"))
	},
	Main: func(ctx *Context) ExitStatus {
		ctx.App().Println("Hello world!", *ctx.MustGet("top").(*int))
		return ExitUsage
	},

	SubCommands: []Command{
		{
			Name:  "secondary",
			Usage: "${fullName} ${shortFlags}",
			SetFlags: func(ctx *Context) {
				ctx.Set("second", ctx.FlagSet().Int("second", 21, "example second-level flag"))
			},
			Main: func(ctx *Context) ExitStatus {
				ctx.App().Println("Secondary world!")
				return ExitCmd
			},
			app: app,
		},
	},

	app: app,
}
var testCmd = &TmplCmdWithSubCmd

// TestFullName ensures that FullName returns expected strings.
func TestFullName(t *testing.T) {
	if res := testCmd.FullName(); res != "test" {
		t.Errorf("Command.FullName: got '%s' expected 'test'", res)
	}

	TmplCmdWithSubCmd.SubCommands[0].parent = &TmplCmdWithSubCmd
	if res := testCmd.SubCommands[0].FullName(); res != "test secondary" {
		t.Errorf("Command.SubCommands[0].FullName: got '%s' expected 'test secondary'", res)
	}
}

// TestCmdNewContext ensures that NewContext returns a properly populated
// Context for both top- and second-level commands.
func TestCmdNewContext(t *testing.T) {
	topCtx := testCmd.NewContext()

	if topCtx.App().Name != "command_test" {
		t.Errorf("Command.NewContext.App.Name: got '%s' expected 'command_test'", topCtx.App().Name)
	}
	if topCtx.Command().Name != "test" {
		t.Errorf("Command.NewContext.Command.Name: got '%s' expected 'test'", topCtx.Command().Name)
	}
	if topCtx.FlagSet().Name() != "test" {
		t.Errorf("Command.NewContext.FlagSet.Name: got '%s' expected 'test'", topCtx.FlagSet().Name())
	}
	if topCtx.Parent() != nil {
		t.Errorf("Command.NewContext.Parent: expected nil got:\n%#v", topCtx.Parent())
	}

	secondCtx := testCmd.SubCommands[0].NewContext()

	if secondCtx.Parent().Name != "test" {
		t.Errorf("Command.NewContext.Parent.Name: got '%s' expected 'test'", secondCtx.Parent().Name)
	}
}

// TestMatch ensures that Match correctly handles calls to top- and second-
// level commands.
func TestMatch(t *testing.T) {
	if res, err := testCmd.Match([]string{"test", "none"}); err != nil {
		t.Error("Command.Match: got error:\n", err)
	} else if res != testCmd {
		t.Errorf("Command.Match: got command '%s' expected 'test'", res.Name)
	}

	if res, err := testCmd.Match([]string{"test", "secondary"}); err != nil {
		t.Error("Command.Match: got error:\n", err)
	} else if res.Name != "secondary" {
		t.Errorf("Command.Match: got command '%s' expected 'secondary'", res.Name)
	}

	if _, err := testCmd.Match([]string{"other"}); err == nil {
		t.Error("Command.Match: expected error with non-existent command")
	} else if !strings.Contains(err.Error(), "does not match") {
		t.Error("Command.Match: got unexpected error message with non-existent command:\n", err)
	}
}

// TestExecute ensures that Execute correctly handles calls to execute top- and
// second-level commands.
func TestExecute(t *testing.T) {
	if status, err := testCmd.Execute([]string{"test", "-top", "19"}); err != nil {
		t.Error("Command.Execute: got error:\n", err)
	} else if status != ExitUsage {
		t.Errorf("Command.Execute: got ExitStatus '%d' expected '%d'", status, ExitUsage)
	} else if !strings.Contains(output.String(), "${name} ${shortFlags}:\n\nExecute") {
		t.Error("Command.Execute: expected output to contain Usage with ExitUsage, got:\n", output.String())
	}

	if status, err := testCmd.SubCommands[0].Execute([]string{"test", "secondary"}); err != nil {
		t.Error("Command.Execute: got error:\n", err)
	} else if status != ExitCmd {
		t.Errorf("Command.Execute: got ExitStatus '%d' expected '%d'", status, ExitCmd)
	}

	if _, err := testCmd.Execute([]string{"test", "---hello"}); err == nil {
		t.Error("Command.Execute: expected error with invalid flags")
	} else if _, ok := err.(*ErrParseFlags); !ok {
		t.Error("Command.Execute: expected error of type *ErrParseFlags with invalid flags:\n", err)
	}
}
