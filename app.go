package shell

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
)

// ErrNoCmd is returned from ExecuteString if the input does not call a valid
// command.
type ErrNoCmd struct {
	Name string
}

// Error implements the error interface for ErrNoCmd.
func (err *ErrNoCmd) Error() string {
	return fmt.Sprintf("App.ExecuteString: command '%s' not found", err.Name)
}

// ErrParseInput is returned from ExecuteString is the input for some reason
// cannot be parsed. Currently occurs only when input is empty.
type ErrParseInput struct {
	Input string
}

// Error implements the error interface for ErrParseInput
func (err *ErrParseInput) Error() string {
	return fmt.Sprintf("App.ExecuteString: failed to parse input '%s'", err.Input)
}

// App is the main structure that makes up a single shell. Through it commands
// are created and managed. App is not intended to be directly created or
// manipulated, instead its methods and NewApp should be utilized.
type App struct {
	// Name is the name of the program. No restrictions are applied. Defaults
	// to filepath.Base(os.Args[0]).
	Name string

	// Commands holds all commands attached to the App.
	Commands []*Command

	// Output controls the destination for general messages emitted by the App.
	Output io.Writer

	// ErrOutput controls the destination for usage and error messages.
	ErrOutput io.Writer

	// Input controls the reader used to fetch user input.
	Input io.ReadCloser
}

// NewApp creates an App and configures its logger. The first argument defines
// the name of the App and defaults to filepath.Base(os.Args[0]) if blank. A
// set of default commands  are also automatically added unless false is passed
// as an argument.
func NewApp(name string, addDefaults bool) *App {
	if name == "" {
		name = filepath.Base(os.Args[0])
	}

	app := &App{
		Name:      name,
		Commands:  make([]*Command, 0),
		Output:    os.Stdout,
		ErrOutput: os.Stderr,
		Input:     os.Stdin,
	}

	if addDefaults {
		for _, cmd := range DefaultCommands {
			if err := app.AddCommand(*cmd); err != nil {
				panic(err)
			}
		}
	}

	return app
}

// Print prints to the App's Output. Arguments are handled in the manner of
// fmt.Print.
func (app *App) Print(a ...interface{}) {
	fmt.Fprint(app.Output, a...)
}

// Printf prints to the App's Output. Arguments are handled in the manner of
// fmt.Printf.
func (app *App) Printf(format string, a ...interface{}) {
	fmt.Fprintf(app.Output, format, a...)
}

// Println prints to the App's Output. Arguments are handled in the manner of
// fmt.Println.
func (app *App) Println(a ...interface{}) {
	fmt.Fprintln(app.Output, a...)
}

// GetByName takes a string and returns a pointer to a command or an error if
// no command by that name exists.
func (app *App) GetByName(name string) (*Command, error) {
	for _, cmd := range app.Commands {
		if name == cmd.Name {
			return cmd, nil
		}
	}

	return nil, fmt.Errorf("App.GetByName: command '%s' does not exist", name)
}

// getDefaults takes a FlagSet and returns a string containing the result of
// PrintDefaults.
func getDefaults(flags *flag.FlagSet) string {
	oldOutput := flags.Output()
	flags.SetOutput(&strings.Builder{})
	flags.PrintDefaults()
	output := flags.Output().(*strings.Builder).String()
	flags.SetOutput(oldOutput)
	return output
}

// getShortDefaults takes a FlagSet and returns a string containing a space-
// separated list of all available flags surrounded by brackets.
func getShortDefaults(flags *flag.FlagSet) string {
	output := &strings.Builder{}
	flags.VisitAll(func(item *flag.Flag) {
		fmt.Fprintf(output, " [-%s]", item.Name)
	})
	return output.String()[1:]
}

// AddCommand takes a Command and adds it to the App. If the command or any of
// its sub-commands are invalid an error is returned.
func (app *App) AddCommand(cmd Command) error {
	if _, err := app.GetByName(cmd.Name); err == nil {
		return fmt.Errorf("App.AddCommand: '%s' already exists", cmd.Name)
	}

	if len(cmd.SubCommands) > 0 {
		// Add default sub-commands
		if cmd.PreventDefaultSubCommands != true {
			for _, def := range DefaultSubCommands {
				switch def.Name {
				case "flags":
					for _, item := range append(cmd.SubCommands, cmd) {
						if item.SetFlags != nil {
							cmd.SubCommands = append(cmd.SubCommands, def)
						}
					}
				default:
					cmd.SubCommands = append(cmd.SubCommands, def)
				}
			}
		}

		for key := range cmd.SubCommands {
			subCmd := &cmd.SubCommands[key]
			// if any sub-commands have name beginning with '-', return an error
			if subCmd.Name != "" && subCmd.Name[0] == '-' {
				return fmt.Errorf("App.AddCommand: sub-commands must not begin with the character '-'")
			}

			// if any sub-commands contain second-level sub-commands, return an error
			if len(subCmd.SubCommands) > 0 {
				return fmt.Errorf("App.AddCommand: '%s' contains more than one level of sub-commands", cmd.Name)
			}

			subCmd.parent = &cmd
		}
	}

	items := append(make([]*Command, 0), &cmd)
	for key := range cmd.SubCommands {
		items = append(items, &cmd.SubCommands[key])
	}

	for _, item := range items {
		if item.Name == "" {
			return fmt.Errorf("App.AddCommand: (sub-)command name cannot be blank")
		}

		spaces := 0
		// Count whitespace in command name
		for _, char := range item.Name {
			if unicode.IsSpace(char) {
				spaces++
			}
		}

		if spaces > 0 {
			return fmt.Errorf("App.AddCommand: (sub-)command name '%s' contains %d disallowed whitespace characters", item.Name, spaces)
		}

		// if any commands are missing Main functions, return an error
		if item.Main == nil {
			return fmt.Errorf("App.AddCommand: 'Main' function for (sub-)command '%s' is nil", item.Name)
		}

		item.app = app

		// Parse templates in Usage field
		item.Usage = strings.ReplaceAll(item.Usage, "${name}", item.Name)
		item.Usage = strings.ReplaceAll(item.Usage, "${fullName}", item.FullName())

		itemCtx := item.NewContext()
		if item.SetFlags != nil {
			item.SetFlags(itemCtx)

			item.Usage = strings.ReplaceAll(item.Usage, "${flags}", getDefaults(itemCtx.FlagSet()))
			item.Usage = strings.ReplaceAll(item.Usage, "${shortFlags}", getShortDefaults(itemCtx.FlagSet()))
		} else {
			item.Usage = strings.ReplaceAll(item.Usage, "${flags}", "")
			item.Usage = strings.ReplaceAll(item.Usage, "${shortFlags}", "")
			item.Usage = strings.TrimSpace(item.Usage)
		}
	}

	app.Commands = append(app.Commands, &cmd)

	return nil
}

// ExecuteString takes what is usually some user input and attempts to execute
// a command based on the input. If no matching command exists an ErrNoCmd is
// returned. If the input string is invalid an ErrParseInput is returned. If a
// command is successfully executed, it's ExitStatus is returned, otherwise
// ExecuteString defaults to ExitCmd. An ErrParseFlags may be returned in event
// of a failure when parsing the input flags.
func (app *App) ExecuteString(input string) (ExitStatus, error) {
	split := strings.Fields(input)
	if len(split) > 0 {
		for _, cmd := range app.Commands {
			if item, err := cmd.Match(split); err == nil {
				// if item has a parent it is a sub-command, pass split from the second string onward
				if item.parent != nil {
					return item.Execute(split[1:])
				}

				return item.Execute(split)
			}
		}

		return ExitCmd, &ErrNoCmd{Name: split[0]}
	}

	return ExitCmd, &ErrParseInput{Input: input}
}

// Main is the App's main loop. It accepts user input infinitely until some
// command returns an ExitStatus of ExitShell. Any errors that occur are not
// propagated back up but rather printed to the App's Output.
func (app *App) Main() ExitStatus {
	app.Println("Welcome to the shell. Type \"help\" for available Commands.")

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
		Stdin:  app.Input,
		Stdout: app.Output,
		Stderr: app.ErrOutput,
	})

	if err != nil {
		panic(fmt.Sprintf("App: got error while initializing readline:\n%s", err))
	}

	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil { // error is io.EOF or readline.ErrInterrupt
			return ExitShell
		}

		// if input is blank, ignore
		if strings.TrimSpace(input) == "" {
			continue
		}

		exitStatus, err := app.ExecuteString(input)

		// switch err type:
		//	is flag parse error => app.print
		//	is no matching command error => print("%s: command not found")
		//	is failed to parse input error => print("failed to parse input")

		if err != nil {
			switch val := err.(type) {
			case *ErrParseFlags:
				app.Printf("%s: failed to parse flags:\n%s", val.Name, val.Err)
			case *ErrNoCmd:
				app.Printf("%s: command not found", val.Name)
			default:
				app.Println(err)
			}
		}

		if exitStatus != ExitCmd {
			return exitStatus
		}
	}
}
