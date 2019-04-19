# shell

shell provides a simple shell interface and an API to register commands and sub-commands for use within it. The goal of this shell package is to provide an API that makes the task of creating a functional custom shell as easy as possible.

For more information and API documentation see [GoDoc](https://godoc.org/github.com/octacian/shell).

## Example

```go
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

app.Main() // Start the main loop.
```
