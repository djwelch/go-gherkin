package gherkin

import (
    "fmt"
    "io"
    "testing"
)

var t *testing.T

// Passed to each step-definition
type World struct {
    regexParams []string
    multiStep []map[string]string
    output io.Writer
    gotAnError bool
}

// Allows World to be used with the go-matchers AssertThat() function.
func (w *World) Errorf(format string, args ...interface{}) {
    w.gotAnError = true
    if w.output != nil {
        if len(args) == 0 {
            fmt.Fprintf(w.output, format)
            if t != nil {
                t.Error(format)
            }
        } else {
            fmt.Fprintf(w.output, format, args)
            if t != nil {
                t.Error(format, args)
            }
        }
    }
}
