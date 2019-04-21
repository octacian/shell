package shell

import (
	"flag"
	"fmt"
)

// ErrParseFlags is returned from Command.Execute is the FlagSet fails to parse.
type ErrParseFlags struct {
	Name string
	Err  error
}

// Error implements the error interface for ErrParseFlags.
func (err *ErrParseFlags) Error() string {
	return fmt.Sprintf("App.Execute: failed to parse flags for '%s':\n%s", err.Name, err.Err)
}

// Command is a top-level command within a shell App. It may contain an
// arbitrary number of sub-command.
type Command struct {
	// Name is required and should be as concise as possible. It may not
	// contain any spaces.
	Name string

	// Synopsis should contain a short description of the command. It usually
	// should not be more than a single sentence.
	Synopsis string

	// Usage should contain a detailed description of the command. There are no
	// limitations to its length. Several sequences are substituted with
	// information relating to the command when found within the usage string:
	// `${name}` is substituted with the name of the command, ${fullName} with
	// the full name of the command (including parent command name if the
	// command is a sub-command), ${flags} with the help information for the
	// command flags as described by/ flag.PrintDefaults, and ${shortFlags} for
	// a short list of all registered flags in the format of [-<flag name>] and
	// separated with spaces.
	Usage string

	// SetFlags should register any flags with the flag.FlagSet available
	// through the Context and store their result within via Context.[Get|Set].
	// SetFlags should not attempt to parse flags since it does not have access
	// to the complete input string.
	SetFlags func(*Context)

	// Main is required and contains the command logic itself. If SetFlags
	// exists, flags will be parsed immediately before Main is called and
	// the results should be accessible via the Context.
	Main func(*Context) ExitStatus

	// SubCommands should contain an arbitrary number of Commands. If the name
	// of a valid sub-command directly follows the name of this command in some
	// user input, the sub-command will be preferred over this Command.
	// Otherwise, this Command will be executed.
	SubCommands []Command

	// PreventDefaultSubCommands controls whether the sub-commands defined
	// within the exported table PreventDefaultSubCommands should be added by
	// default. If no sub-commands are defined in the SubCommands array, this
	// option is disregarded. If left blank, default sub-commands are added.
	PreventDefaultSubCommands bool

	// parent holds a pointer to the parent Command if this Command is in fact
	// a sub-command.
	parent *Command

	// app holds a pointer to the App which this Command is a part of.
	app *App
}

// FullName returns the full name of the command, checking if it has a parent
// and if so prepending it to its own name.
func (cmd *Command) FullName() string {
	if cmd.parent != nil {
		return fmt.Sprintf("%s %s", cmd.parent.Name, cmd.Name)
	}

	return cmd.Name
}

// NewContext returns an empty context prepared for this command.
func (cmd *Command) NewContext() *Context {
	flagSet := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	flagSet.SetOutput(cmd.app.ErrOutput)

	return NewContext(cmd.app, cmd, flagSet, cmd.parent)
}

// GetSubCommand attempts to fetch a sub-command by name, returning a pointer
// to the sub-command if successful and an error if it does not exist.
func (cmd *Command) GetSubCommand(name string) (*Command, error) {
	for _, subCmd := range cmd.SubCommands {
		if name == subCmd.Name {
			return &subCmd, nil
		}
	}

	return nil, fmt.Errorf("Command.GetSubCommand: sub-command '%s' does not exist", name)
}

// Match takes an array of strings, usually representing some user input
// retrieved from the shell loop. If the input does not call for this command
// an error is returned, otherwise Match checks if the input calls for a sub-
// command, returning either it or this Command if no match is found.
func (cmd *Command) Match(input []string) (*Command, error) {
	if input[0] == cmd.Name {
		if len(cmd.SubCommands) > 0 && len(input) > 1 && input[1][1] != '-' {
			if subCmd, err := cmd.GetSubCommand(input[1]); err == nil {
				return subCmd, nil
			}
		}

		return cmd, nil
	}

	return nil, fmt.Errorf("Command.Match: input does not match command '%s'", cmd.Name)
}

// Execute takes an array of strings, usually representing some user input
// retrieved from the shell loop. It then executes this Command, first parsing
// the input for flags. If an error occurs while parsing flags, it is returned.
func (cmd *Command) Execute(input []string) (ExitStatus, error) {
	ctx := cmd.NewContext()

	// if SetFlags function has been set, call it
	if cmd.SetFlags != nil {
		cmd.SetFlags(ctx)
	}

	// Parse flagSet
	if err := ctx.FlagSet().Parse(input[1:]); err != nil {
		return ExitCmd, &ErrParseFlags{Name: cmd.Name, Err: err}
	}

	return cmd.Main(ctx), nil
}
