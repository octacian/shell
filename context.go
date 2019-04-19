package shell

import (
	"flag"
	"fmt"
)

// Context is a type that is passed to each handler when a Command is executed
// and allows arbitrary information to be shared between handlers besides
// providing access to specifics about the command itself. Context may be
// created manually due to its simplicity, but it is often helpful to utilize
// NewContext.
type Context struct {
	// app is the App to which the Context belongs.
	app *App

	// command is the Command for which the Context is acting.
	command *Command

	// flagSet is the flag.FlagSet for use with command handler functions.
	flagSet *flag.FlagSet

	// parent is only intended to be used when the Context is acting for a sub-
	// command and should hold a pointer to the parent Command.
	parent *Command

	// values must be initialized as a slice and is used to perform CRUD
	// operations on data passed through the Context.
	values map[string]interface{}
}

// NewContext creates a new context given an App, a Command, a flag.FlagSet,
// and optionally a parent Command.
func NewContext(app *App, command *Command, flagSet *flag.FlagSet, parent *Command) *Context {
	return &Context{
		app:     app,
		command: command,
		flagSet: flagSet,
		parent:  parent,
		values:  make(map[string]interface{}),
	}
}

// App returns the connected shell App. Warning: if none exists a nil pointer
// will be returned.
func (context *Context) App() *App {
	return context.app
}

// Command returns the Command for which the Context is acting. Warning: if
// none exists a nil pointer will be returned.
func (context *Context) Command() *Command {
	return context.command
}

// FlagSet returns the flag.FlagSet for use with command handler functions.
// Warning: if none exists a nil pointer will be returned.
func (context *Context) FlagSet() *flag.FlagSet {
	return context.flagSet
}

// Parent returns the parent Command. Warning: if none exists a nil pointer
// will be returned.
func (context *Context) Parent() *Command {
	return context.parent
}

// Get takes a string and returns its value or an error if the key does not
// exist.
func (context *Context) Get(name string) (interface{}, error) {
	if val, ok := context.values[name]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("Context.Get: value '%s' does not exist", name)
}

// ShouldGet does the same as Get but returns nil if the key does not exist.
func (context *Context) ShouldGet(name string) interface{} {
	return context.values[name]
}

// MustGet does the same as Get but panics if an error is returned.
func (context *Context) MustGet(name string) interface{} {
	val, err := context.Get(name)
	if err != nil {
		panic(err)
	}

	return val
}

// Set takes a string and an interface and sets a value in the context.
func (context *Context) Set(name string, value interface{}) {
	context.values[name] = value
}

// Delete takes a string and delete the value stored at that index. No error is
// returned regardless of whether a deletion actually occurs.
func (context *Context) Delete(name string) {
	delete(context.values, name)
}
