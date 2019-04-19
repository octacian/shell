package shell

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

// TODO: Allow data return by the function to be handled. Channels?
// panicked takes a simple function to execute and returns an error containing
// the data passed to panic from within the function, and nil if no panic
// occurred.
func panicked(fn func()) error {
	ch := make(chan error)
	go func() {
		defer func() {
			// if function didn't panic, return nil
			if r := recover(); r == nil {
				ch <- nil
			} else { // else, return error
				switch r.(type) {
				case error:
					ch <- r.(error)
				case string:
					ch <- errors.New(r.(string))
				default:
					ch <- nil
				}
			}
		}()

		fn()
	}()

	return <-ch
}

// TestNewContext ensures that a parent context is also created when a parent
// command is provided and that all other methods provide accurate results.
func TestNewContext(t *testing.T) {
	ctx := NewContext(&App{}, &Command{Name: "child"}, flag.NewFlagSet("TestNewContext",
		flag.PanicOnError), &Command{Name: "parent"})

	if ctx.App() == nil {
		t.Error("NewContext.App: got nil expected *App")
	}

	if res := ctx.Command(); res == nil {
		t.Error("NewContext.Command: got nil expected *Command")
	} else if res.Name != "child" {
		t.Errorf("NewContext.Command: got name '%s' expected 'child'", res.Name)
	}

	if res := ctx.FlagSet(); res == nil {
		t.Error("NewContext.FlagSet: got nil expected *flag.FlagSett")
	} else if res.Name() != "TestNewContext" {
		t.Errorf("NewContext.FlagSet: got name '%s' expected 'TestNewContext'", res.Name())
	}

	if res := ctx.Parent(); res == nil {
		t.Error("NewContext.Parent: got nil parent expected *Command")
	} else if res.Name != "parent" {
		t.Errorf("NewContext.Parent: got command name '%s' expected 'parent'", res.Name)
	}
}

// TestContextStore ensures get/set/delete operations perform as expected.
func TestContextStore(t *testing.T) {
	ctx := &Context{values: make(map[string]interface{})}

	if _, err := ctx.Get("foo"); err == nil {
		t.Error("Context.Get: expected error with non-existent key")
	} else if !strings.Contains(err.Error(), "does not exist") {
		t.Error("Context.Get: got unexpected error message with non-existent key:\n", err)
	}

	if res := ctx.ShouldGet("foo"); res != nil {
		t.Errorf("Context.ShouldGet: got '%v' expected 'nil' with non-existent key", res)
	}

	if err := panicked(func() { ctx.MustGet("foo") }); err == nil {
		t.Error("Context.MustGet: expected panic with non-existent key")
	} else if !strings.Contains(err.Error(), "does not exist") {
		t.Error("Context.MustGet: got unexpected error message with non-existent key:\n", err)
	}

	ctx.Set("foo", "bar")

	if res, err := ctx.Get("foo"); err != nil {
		t.Error("Context.Get: got error:\n", err)
	} else if res != "bar" {
		t.Errorf("Context.Get: got '%v' expected 'bar'", res)
	}

	if res := ctx.ShouldGet("foo"); res != "bar" {
		t.Errorf("Context.ShouldGet: got '%v' expected 'bar'", res)
	}

	// TODO: Ensure no panic as well as checking result.
	if res := ctx.MustGet("foo"); res != "bar" {
		t.Errorf("Context.MustGet: got '%v' expected 'bar'", res)
	}

	ctx.Delete("foo")

	if res := ctx.ShouldGet("foo"); res != nil {
		t.Errorf("Context.Delete: got '%v' after delete expected 'nil'", res)
	}
}
