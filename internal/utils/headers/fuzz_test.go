package headers

import (
	"testing"
)

func FuzzParseHeaders(f *testing.F) {
	seeds := []string{
		"User-Agent: Test",
		"Content-Type: application/json",
		"Invalid",
		"NoColon",
		":EmptyKey",
		"EmptyVal:",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Pass slice of strings
		_ = ParseHeaders([]string{input})
	})
}
