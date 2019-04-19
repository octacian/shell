package shell

import (
	"testing"
)

// WithSubCommands runs a function providing an app with a 'test' command that
// includes one of its own sub-commands as well as all default sub-commands.
func WithSubCommands(t *testing.T, name string, fn func(*App)) {
	app := NewApp(name, true)
	if err := app.AddCommand(TmplCmdWithSubCmd); err != nil {
		t.Fatalf("App(Name: %s).AddCommand: got error:\n%s", name, err)
	} else {
		fn(app)
	}
}

// TestExitCommand tests the default top-level exit command.
func TestExitCommand(t *testing.T) {
	app := NewApp("TestExitCommand", true)

	MainInputWithStatus(t, app, "exit command with disallowed arguments", "exit hello", ExitCmd, "Usage: exit")
	MainInputWithStatus(t, app, "exit command with -shell-only flag", "exit -shell-only", ExitShell)
	MainInputWithStatus(t, app, "exit command with no flags", "exit", ExitAll)
}

// TestHelpCommand tests the default top-level help command.
func TestHelpCommand(t *testing.T) {
	app := NewApp("TestHelpCommand", true)

	if err := app.AddCommand(TmplSimpleCmd); err != nil {
		t.Error("App.AddCommand: got error:\n", err)
	} else {
		MainInput(t, app, "help with no arguments", "help", "Available commands", "exit shell", "testing command")
		MainInput(t, app, "help for 'test' command including Usage string", "help test", "test", "testing command", "tests stuff")
	}

	MainInput(t, app, "help for 'exit' command", "help exit", "exit", "exit shell")
	MainInput(t, app, "help for non-existent 'nothing' command", "help nothing", "command not found")
	MainInput(t, app, "help with too many arguments", "help exit test", "Usage: help")
}

// TestCommandsSubCommand tests the default second-level commands command.
func TestCommandSubCommand(t *testing.T) {
	WithSubCommands(t, "TestCommandSubCommand", func(app *App) {
		MainInput(t, app, "commands sub-command", "test commands", "commands", "flags", "help", "secondary")
		MainInput(t, app, "commands with too many arguments", "test commands something", "Print a list")
	})
}

// TestFlagsSubCommand tests the default second-level flags command.
func TestFlagsSubCommand(t *testing.T) {
	WithSubCommands(t, "TestFlagsSubCommand", func(app *App) {
		MainInput(t, app, "flags with too many arguments", "test flags secondary none", "flags [<sub-command>]:")
		MainInput(t, app, "flags with no arguments", "test flags", "example top-level")
		MainInput(t, app, "flags for 'secondary' sub-command", "test flags secondary", "example second-level")
		MainInput(t, app, "flags for non-existent 'nothing' sub-command", "test flags nothing", "test nothing", "not found")
	})
}

// TestHelpSubCommand tests the default second-level help command.
func TestHelpSubCommand(t *testing.T) {
	WithSubCommands(t, "TestHelpSubCommand", func(app *App) {
		MainInput(t, app, "help with too many arguments", "test help secondary none", "help [<sub-command>]:")
		MainInput(t, app, "help with no arguments", "test help", "Usage: test <sub-command>", "secondary",
			"commands", "list all", "flags", "describe all")
		MainInput(t, app, "help for 'flags' sub-command", "test help flags", "flags [<sub-command>]:")
		MainInput(t, app, "help for non-existant 'nothing' sub-command", "test help nothing", "test nothing", "not found")
	})
}
