package ui

import "fmt"

// HintedError is an error that includes a user-friendly hint
type HintedError struct {
	Err  error
	Hint string
}

func (e *HintedError) Error() string {
	return e.Err.Error()
}

func (e *HintedError) Unwrap() error {
	return e.Err
}

// NewHintedError creates a new error with a hint
func NewHintedError(err error, hint string) error {
	return &HintedError{
		Err:  err,
		Hint: hint,
	}
}

// WrapWithHint wraps an existing error with a hint
func WrapWithHint(err error, hint string) error {
	if err == nil {
		return nil
	}
	return &HintedError{
		Err:  err,
		Hint: hint,
	}
}

// PrintError beautifully prints an error and its hint if available
func PrintError(err error) {
	if err == nil {
		return
	}

	fmt.Println(StyledError("Error: " + err.Error()))

	if hintedErr, ok := err.(*HintedError); ok && hintedErr.Hint != "" {
		fmt.Println(StyledHint(hintedErr.Hint))
	}
}
