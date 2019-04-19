/*
Package shell provides a simple shell interface and an API to register
commands and sub-commands for use within it.

Basics

Create a new app and add a command:

	// The first argument is the app name and the second controls whether
	// or not the default commands help and exit are added.
	app := shell.NewApp("MyApp", true)

	if err := app.AddCommand(Command{
		Name: "test",
		Synopsis: "run some tests",
		Usage: "${name} ${shortFlags}:

	Do some things and run some tests and more detailed information.

	${flags}",
		SetFlags: func(ctx *shell.Context) {
			ctx.Set("top", ctx.FlagSet().Int("top", 12, "example top-level flag"))
		},
		Main: func(ctx *shell.Context) shell.ExitStatus {
			ctx.App().Println("Hello world!", *ctx.MustGet("top").(*int))
			return ExitCmd
		},
		SubCommands: []Command{
			{
				Name: "secondary",
				SetFlags: func(ctx *Context) {
					ctx.Set("second", ctx.FlagSet().Int("second", 21, "example second-level flag"))
				},
				Main: func(ctx *Context) ExitStatus {
					ctx.App().Println("Hello world from a sub-command!")
					return ExitCmd
				},
			},
		}
	})

Start the main loop:

	app.Main()

And you're all set!
*/
package shell
