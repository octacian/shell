package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var blankSetFlagsFunc = func(ctx *Context) {
	ctx.FlagSet().Bool("test", true, "testing bool")
}
var blankMainFunc = func(_ *Context) ExitStatus { return ExitCmd }

var TmplSimpleCmd = Command{
	Name:     "test",
	Synopsis: "testing command",
	Usage:    "tests stuff",
	Main: func(ctx *Context) ExitStatus {
		ctx.App().Println("Hello world from test command!")
		return ExitCmd
	},
}

var TmplBlankSubCmd = Command{
	Name: "second",
	Main: blankMainFunc,
}

// MainInput takes a test state, an app, a message, an input string, and a
// variable number of substring arguments. The output of running the input
// string via App.Main is checked for each substring, and t.Error is called
// with the message "expected substring '<substring>' while testing <message>"
// if any substring is missing from the output string. Any empty substrings are
// ignored.
func MainInput(t *testing.T, app *App, msg, in string, substr ...string) {
	MainInputWithStatus(t, app, msg, in, 0, substr...)
}

// MainInputWithStatus does the same as MainInput but takes an additional
// argument holding the expected ExitStatus before the variable number of
// substring arguments and checks that it was returned. If the exit argument
// is 0, the ExitStatus is not tested.
func MainInputWithStatus(t *testing.T, app *App, msg, in string, exit ExitStatus, substr ...string) {
	oldOutput := app.Output
	output := &strings.Builder{}
	errOutput := &strings.Builder{}
	app.Output = output
	app.ErrOutput = errOutput

	output.Reset()
	errOutput.Reset()
	app.Input = ioutil.NopCloser(strings.NewReader(in))
	status := app.Main()

	for _, str := range substr {
		if str != "" && !strings.Contains(output.String(), str) && !strings.Contains(errOutput.String(), str) {
			t.Errorf("App.Main: expected substring '%s' while testing %s, got output:\n%s\nand error output:\n%s",
				str, msg, output.String(), errOutput.String())
		}
	}

	if exit != 0 && status != exit {
		t.Errorf("App.Main: expected ExitStatus '%d' while testing %s, got '%d'", exit, msg, status)
	}

	app.Output = oldOutput
}

// TestNewApp checks that App defaults are as expected with default commands
// and that GetByName gives expected results.
func TestNewApp(t *testing.T) {
	app := NewApp("TestNewApp", true)

	if app.Output != os.Stdout {
		t.Fatal("App.Writer: expected default of `os.Stdout`")
	}

	if app.ErrOutput != os.Stderr {
		t.Fatal("App: expected ErrOutput default of `os.Stderr`")
	}

	if app.Input != os.Stdin {
		t.Fatal("App: expected Input default of `os.Stdin`")
	}

	if len(app.Commands) != 2 {
		t.Errorf("App: got %d items in commands expected 2", len(app.Commands))
	}

	if res, err := app.GetByName("help"); err != nil {
		t.Fatal("App.GetByName: got error:\n", err)
	} else if res.Name != "help" {
		t.Fatalf("App.GetByName: got command '%s' expected 'help'", res.Name)
	}

	if nameless := NewApp("", false); nameless.Name == "" {
		t.Errorf("App.Name: got blank expected default of `filepath.Base(os.Args[0])`")
	}

	if nodefaults := NewApp("NoDefaults", false); len(nodefaults.Commands) != 0 {
		t.Errorf("App: got %d items in commands expected 0", len(app.Commands))
	}
}

// TestAppPrint runs a few very simple tests to ensure that the App's printing
// methods work as expected.
func TestAppPrint(t *testing.T) {
	app := NewApp("TestAppPrint", false)
	app.Output = &strings.Builder{}

	test := func(str string, fn func()) {
		output := app.Output.(*strings.Builder)
		output.Reset()
		fn()
		if output.String() != str {
			t.Errorf("App.Print[f|ln]: got output '%s' expected '%s'", output.String(), str)
		}
	}

	test("Hello from Print!\n", func() { app.Print("Hello from Print!\n") })
	test("Hello from Printf!\n", func() { app.Printf("Hello from %s!\n", "Printf") })
	test("Hello from Println!\n", func() { app.Println("Hello from Println!") })
}

// TestNewAppWithoutDefaults checks that no default commands are added when
// false is passed as the first argument to NewApp.
func TestNewAppWithoutDefaults(t *testing.T) {
	app := NewApp("TestNewAppWithoutDefaults", false)

	if len(app.Commands) != 0 {
		t.Fatalf("App: got %d items in commands expected 0", len(app.Commands))
	}
}

// TestBadAddCommand checks that the correct errors are returned with a number
// of invalid commands.
func TestBadAddCommand(t *testing.T) {
	app := NewApp("TestBadAddCommand", true)

	expectError := func(cmd *Command, msg string, substr string) {
		if err := app.AddCommand(*cmd); err == nil {
			t.Errorf("App.AddCommand: expected error with %s", msg)
		} else if !strings.Contains(err.Error(), substr) {
			t.Errorf("App.AddCommand: expected error message with %s to contain substring '%s' got:\n%s", msg, substr, err)
		}
	}

	subCmdTmpl := func(subCmd *Command, msg string, substr string) {
		expectError(&Command{
			Name:        "test",
			Main:        blankMainFunc,
			SubCommands: []Command{*subCmd},
		}, msg, substr)
	}

	subCmdMainTmpl := func(subCmd *Command, msg string, substr string) {
		subCmd.Main = blankMainFunc
		subCmdTmpl(subCmd, msg, substr)
	}

	expectError(&Command{Name: "help"}, "duplicate command", "already exists")

	expectError(&Command{Name: ""}, "blank command name", "cannot be blank")
	expectError(&Command{Name: "contains spaces"}, "whitespace in command name", "disallowed whitespace")
	expectError(&Command{Name: "test"}, "missing main function", "function for (sub-)command")

	subCmdTmpl(&Command{Name: ""}, "blank sub-command name", "cannot be blank")
	subCmdTmpl(&Command{Name: "contains spaces"}, "whitespace in sub-command name", "disallowed whitespace")
	subCmdTmpl(&Command{Name: "test"}, "sub-command missing main function", "function for (sub-)command")

	subCmdMainTmpl(&Command{Name: "-no"}, "'-' at start of sub-command name", "must not begin with")
	subCmdMainTmpl(&Command{Name: "second-level", SubCommands: []Command{
		{Name: "third-level"},
	}}, "two levels of sub-commands", "more than one level")
}

// TestWorkingAddCommand ensures that no errors are returned with valid
// command structures.
func TestWorkingCommand(t *testing.T) {
	app := NewApp("TestWorkingCommand", false)

	cmdTmpl := Command{
		SetFlags:    blankSetFlagsFunc,
		Main:        blankMainFunc,
		SubCommands: []Command{TmplBlankSubCmd},
	}

	addCommand := func(name string, defaults interface{}) error {
		cmd := cmdTmpl
		cmd.Name = name
		if defaults != nil {
			cmd.PreventDefaultSubCommands = defaults.(bool)
		}
		if err := app.AddCommand(cmd); err != nil {
			t.Error("App.AddCommand: got error with valid command structue:\n", err)
			return err
		}
		return nil
	}

	if addCommand("no-defaults", true) == nil {
		if cmd, _ := app.GetByName("no-defaults"); len(cmd.SubCommands) != 1 {
			t.Errorf("App.AddCommand: got %d sub-commands with command 'no-defaults' expected 1", len(cmd.SubCommands))
		}
	}

	if addCommand("with-defaults", false) == nil {
		if cmd, _ := app.GetByName("with-defaults"); len(cmd.SubCommands) != len(DefaultSubCommands)+1 {
			t.Errorf("App.AddCommand: got %d sub-commands with command 'with-defaults' expected %d",
				len(cmd.SubCommands), len(DefaultSubCommands)+1)
		}
	}

	if addCommand("blank-defaults", nil) == nil {
		if cmd, _ := app.GetByName("blank-defaults"); len(cmd.SubCommands) != len(DefaultSubCommands)+1 {
			t.Errorf("App.AddCommand: got %d sub-commands with command 'blank-defaults' expected %d",
				len(cmd.SubCommands), len(DefaultSubCommands)+1)
		}
	}

	if err := app.AddCommand(Command{
		Name:        "no-setflags",
		Main:        blankMainFunc,
		SubCommands: []Command{TmplBlankSubCmd},
	}); err != nil {
		t.Error("App.AddCommand: got error with no SetFlags function:\n", err)
	} else if cmd, err := app.GetByName("no-setflags"); err != nil {
		t.Error("App.GetByName: got error fetching 'no-setflags':\n", err)
	} else if _, err := cmd.GetSubCommand("flags"); err == nil {
		t.Error("App.AddCommand: found 'flags' sub-command with nil SetFlags function")
	}
}

// TestTemplateReplacement tests whether templates in the Usage field of
// commands is properly replaced.
func TestTemplateReplacement(t *testing.T) {
	app := NewApp("TestTemplateReplacement", false)

	if err := app.AddCommand(TmplCmdWithSubCmd); err != nil {
		t.Error("App.AddCommand: got error:\n", err)
	} else if res, err := app.GetByName("test"); err != nil {
		t.Error("App.GetByName: got error:\n", err)
	} else {
		if !strings.Contains(res.Usage, "test [-top]:") {
			t.Errorf("App.Command.Usage: expected substring 'test [-top]:' got:\n%s", res.Usage)
		}

		if !strings.Contains(res.Usage, "example top-level flag") {
			t.Errorf("App.Command.Usage: expected substring 'example top-level flag' got:\n%s", res.Usage)
		}

		if !strings.Contains(res.SubCommands[0].Usage, "test secondary [-second]") {
			t.Errorf("App.Command.SubCommand.Usage: expected substring 'test secondary [-second]' got:\n%s",
				res.SubCommands[0].Usage)
		}
	}
}

// TestExecuteString ensures that the right command is executed with a variety
// of inputs.
func TestExecuteString(t *testing.T) {
	app := NewApp("TestExecuteString", false)
	app.Output = &strings.Builder{}
	app.ErrOutput = &strings.Builder{}

	if err := app.AddCommand(TmplSimpleCmd); err != nil {
		t.Error("App.AddCommand: got error:\n", err)
	} else {
		if status, err := app.ExecuteString("test"); err != nil {
			t.Error("App.ExecuteString: got error:\n", err)
		} else if status != ExitCmd {
			t.Errorf("App.ExecuteString: got ExitStatus %d expected %d", status, ExitCmd)
		} else if !strings.Contains(app.Output.(*strings.Builder).String(), "Hello world from test command!") {
			t.Error("App.ExecuteString: expected output to contain 'Hello world from test command!' got:\n",
				app.Output.(*strings.Builder).String())
		}

		if _, err := app.ExecuteString("nothing"); err == nil {
			t.Error("App.ExecuteString: expected error with non-existent command")
		} else if _, ok := err.(*ErrNoCmd); !ok {
			t.Error("App.ExecuteString: expected error of type *ErrNoCmd with non-existent command:\n", err)
		}

		if _, err := app.ExecuteString(""); err == nil {
			t.Error("App.ExecuteString: expected error with empty input")
		} else if _, ok := err.(*ErrParseInput); !ok {
			t.Error("App.ExecuteString: expected error of type *ErrParseInput with empty input:\n", err)
		}

		if _, err := app.ExecuteString("test ---hello"); err == nil {
			t.Error("App.ExecuteString: expected error with invalid flags")
		} else if _, ok := err.(*ErrParseFlags); !ok {
			t.Error("App.ExecuteString: expected error of type *ErrParseFlags with invalid flags:\n", err)
		}
	}
}

// TestMain ensures that expected output and ExitStatus is received from Main.
func TestMain(t *testing.T) {
	app := NewApp("TestMain", false)

	if err := app.AddCommand(TmplSimpleCmd); err != nil {
		t.Error("App.AddCommand: got error:\n", err)
	} else {
		MainInput(t, app, "'test' command", "test\n", "Hello world from test command!")
		MainInput(t, app, "'test' with invalid flags", "test ---hello\n", "failed to parse flags")
		MainInput(t, app, "with a non-existent command", "nothing\n", "command not found")
		MainInput(t, app, "with empty input", "\n")
	}

	exitCmd := TmplSimpleCmd
	exitCmd.Name = "exit"
	exitCmd.Main = func(_ *Context) ExitStatus {
		return ExitShell
	}

	if err := app.AddCommand(exitCmd); err != nil {
		for _, cmd := range app.Commands {
			fmt.Println("command:", cmd.Name)
		}
		fmt.Printf("exit: %#v\n", app.Commands[0])
		t.Error("App.AddCommand: got error:\n", err)
	} else {
		app.Input = ioutil.NopCloser(strings.NewReader("exit"))
		if status := app.Main(); status != ExitShell {
			t.Errorf("App.Main: got ExitStatus '%d' expected '%d'", status, ExitShell)
		}
	}
}
