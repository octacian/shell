package shell

import (
	"fmt"
	"sort"
	"strings"
)

// ExitStatus is returned from Command handlers to instruct the program how to
// react to its completion.
type ExitStatus int

const (
	// ExitCmd exits only the command.
	ExitCmd ExitStatus = iota

	// ExitShell exits the shell's infinite loop, but does not itself trigger
	// the entire program to exit.
	ExitShell

	// ExitAll exits not only the shell loop, but also the entire program. It
	// is, however, left up to the enclosing program to respect this.
	ExitAll
)

// DefaultCommands defines the following top-level commands: help and exit.
var DefaultCommands = []*Command{
	{
		Name:     "exit",
		Synopsis: "exit shell",
		SetFlags: func(ctx *Context) {
			ctx.Set("flagOnlyShell", ctx.FlagSet().Bool("shell-only", false, "exit only the shell, "+
				"returning to the main program"))
		},
		Main: func(ctx *Context) ExitStatus {
			// if any arguments were provided, print usage
			if ctx.FlagSet().NArg() > 0 {
				ctx.App().Println("Usage: exit [OPTIONS]")
				ctx.FlagSet().PrintDefaults()
				return ExitCmd
			}

			if *ctx.ShouldGet("flagOnlyShell").(*bool) {
				return ExitShell
			}

			return ExitAll
		},
	},
	{
		Name:     "help",
		Synopsis: "list existing commands and their synopsis",
		Main: func(ctx *Context) ExitStatus {
			switch ctx.FlagSet().NArg() {
			case 0:
				list := make([]string, 0)
				for _, command := range ctx.App().Commands {
					if command.Name == "help" { // Ignore help command
						continue
					}

					list = append(list, fmt.Sprintf("\t%s\t\t%s\n", command.Name, command.Synopsis))
				}
				sort.Strings(list)
				ctx.App().Printf("Available commands:\n%s", strings.Join(list, ""))
			case 1:
				requested, err := ctx.App().GetByName(ctx.FlagSet().Arg(0))
				if err != nil {
					ctx.App().Printf("%s: command not found\n", ctx.FlagSet().Arg(0))
					return ExitCmd
				}

				ctx.App().Printf("%s\t%s\n", requested.Name, requested.Synopsis)

				if requested.Usage != "" {
					ctx.App().Printf("\n%s", requested.Usage)
				}
			default: // if more than 1 argument was provided, print usage
				ctx.App().Println("Usage: help [OPTIONS]")
				ctx.FlagSet().PrintDefaults()
			}

			return ExitCmd
		},
	},
}

// DefaultSubCommands defines the following sub-commands: commands, flags, and
// help.
var DefaultSubCommands = []Command{
	{
		Name:     "commands",
		Synopsis: "list all sub-command names",
		Usage: `${name}:
Print a list of all sub-commands.`,
		Main: func(ctx *Context) ExitStatus {
			if ctx.FlagSet().NArg() > 0 {
				ctx.App().Println(ctx.Command().Usage)
				return ExitCmd
			}

			for _, subCmd := range ctx.Parent().SubCommands {
				ctx.App().Println(subCmd.Name)
			}

			return ExitCmd
		},
	},
	{
		Name:     "flags",
		Synopsis: "describe all known top-level flags",
		Usage: `${name} [<sub-command>]:
With an argument, print all flags of <sub-command>. Else, print a
description of all known top-level flags. (The basic help information only
discusses the most generally important top-level flags.)`,
		Main: func(ctx *Context) ExitStatus {
			flags := ctx.FlagSet()

			if flags.NArg() > 1 {
				ctx.App().Println(ctx.Command().Usage)
				return ExitCmd
			}

			var reqCmd *Command
			if flags.NArg() == 0 {
				reqCmd = ctx.Parent()
			} else {
				for _, cmd := range ctx.Parent().SubCommands {
					if flags.Arg(0) == cmd.Name {
						reqCmd = &cmd
						break
					}
				}

				// if no command was found, print error
				if reqCmd == nil {
					ctx.App().Printf("%s %s: sub-command not found", ctx.Parent().Name, flags.Arg(0))
					return ExitCmd
				}
			}

			reqCtx := reqCmd.NewContext()
			// if setFlags function is provided, call it
			if reqCmd.SetFlags != nil {
				reqCmd.SetFlags(reqCtx)
			}

			reqCtx.FlagSet().PrintDefaults()

			return ExitCmd
		},
	},
	{
		Name:     "help",
		Synopsis: "describe sub-commands and their syntax",
		Usage: `${name} [<sub-command>]:
With an argument, prints detailed information on the use of the specified
sub-command. With no argument, prints a list of all commands and a brief
description of each.`,
		Main: func(ctx *Context) ExitStatus {
			parent := ctx.Parent()

			switch ctx.FlagSet().NArg() {
			case 0:
				ctx.App().Printf("Usage: %s <sub-command> <sub-command args>\n\n"+
					"Sub-commands:\n", parent.Name)

				list := make([]string, 0)
				for _, subCmd := range parent.SubCommands {
					if subCmd.Name == "help" { // Ignore help sub-command
						continue
					}

					list = append(list, fmt.Sprintf("\t%s\t\t%s\n", subCmd.Name, subCmd.Synopsis))
				}
				sort.Strings(list)
				ctx.App().Println(strings.Join(list, ""))
			case 1:
				var reqCmd *Command
				for _, cmd := range parent.SubCommands {
					if ctx.FlagSet().Arg(0) == cmd.Name {
						reqCmd = &cmd
						break
					}
				}

				// if no command was found, print error
				if reqCmd == nil {
					ctx.App().Printf("%s %s: sub-command not found", parent.Name, ctx.FlagSet().Arg(0))
					return ExitCmd
				}

				ctx.App().Println(reqCmd.Usage)
			default:
				ctx.App().Println(ctx.Command().Usage)
			}

			return ExitCmd
		},
	},
}
