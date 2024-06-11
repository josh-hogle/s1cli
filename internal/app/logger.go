package app

import (
	"io"

	"github.com/rs/zerolog"
)

// FilteredLevelWriterConditionFn is called to determine whether or not the given record should be logged.
type FilteredLevelWriterConditionFn func(level zerolog.Level) bool

// FilteredLevelWriterCondition holds a single conditional function to execute.
type FilteredLevelWriterCondition struct {
	// unexported variables
	fn FilteredLevelWriterConditionFn
}

// NewFilteredLevelWriterCondition creates a new FilteredLevelWriterCondition object.
func NewFilteredLevelWriterCondition(fn FilteredLevelWriterConditionFn) *FilteredLevelWriterCondition {
	return &FilteredLevelWriterCondition{
		fn: fn,
	}
}

// And requires this handler condition AND the given function to be true in order to log a record.
//
// Note if either the function stored in this object or the function passed are nil, this condition
// will always return false.
func (c *FilteredLevelWriterCondition) And(fn FilteredLevelWriterConditionFn) *FilteredLevelWriterCondition {
	return &FilteredLevelWriterCondition{
		fn: func(level zerolog.Level) bool {
			if c.fn != nil && fn != nil {
				return c.fn(level) && fn(level)
			}
			return false
		},
	}
}

// Fn returns the actual function associated with the condition that will determine whether or not to log a record.
func (c *FilteredLevelWriterCondition) Fn() FilteredLevelWriterConditionFn {
	return c.fn
}

// Or requires this handler condition OR the given function to be true in order to log a record.
//
// Note if the function stored in this object and the function passed are both nil, this condition
// will always return false.
func (c *FilteredLevelWriterCondition) Or(fn FilteredLevelWriterConditionFn) *FilteredLevelWriterCondition {
	return &FilteredLevelWriterCondition{
		fn: func(level zerolog.Level) bool {
			if c.fn != nil && fn != nil {
				return c.fn(level) || fn(level)
			}
			if c.fn != nil && fn == nil {
				return c.fn(level)
			}
			if c.fn == nil && fn != nil {
				return fn(level)
			}
			return false // nil || nil = false
		},
	}
}

// FilteredLevelWriter filters messages written to the writer based on one or more conditions.
type FilteredLevelWriter struct {
	io.Writer

	// unexported variables
	cond []*FilteredLevelWriterCondition
}

// NewFilteredLevelWriter returns a new FilteredLevelWriter object.
func NewFilteredLevelWriter(w io.Writer, cond []*FilteredLevelWriterCondition) *FilteredLevelWriter {
	return &FilteredLevelWriter{
		Writer: w,
		cond:   cond,
	}
}

// Write writes to the underlying Writer.
func (w *FilteredLevelWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

// WriteLevel will only write the message if all of the filter conditions evaluate to true.
func (fw *FilteredLevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	for _, c := range fw.cond {
		if c == nil || c.fn == nil || !c.fn(level) {
			return len(p), nil
		}
	}
	return fw.Writer.Write(p)
}
