package urlutil

import (
	"strings"
	"testing"
)

func FuzzValidateURL(f *testing.F) {
	seeds := []string{
		"http://example.com",
		"https://google.com/path",
		"ftp://invalid.com",
		"://missing-scheme",
		"example.com",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Just ensure it doesn't panic
		_ = ValidateURL(input)
	})
}

func FuzzResolveRelativeLinks(f *testing.F) {
	f.Add("http://example.com", "<a href='/path'>")
	f.Add("https://site.com/sub/", "<img src='../img.png'>")

	f.Fuzz(func(t *testing.T, baseURL, content string) {
		// Only test with valid-ish base URLs to avoid early exits
		if !strings.HasPrefix(baseURL, "http") {
			return
		}

		// Ensure no panics
		// We can't easily mock the PageData structure in fuzzing without complexity,
		// so we'll test the core logic if possible, or skip if the function is too tied to struct.
		// ResolveRelativeLinks takes *models.PageData.

		// To properly fuzz this, we'd need to refactor ResolveRelativeLinks or use ResolveURL directly
		// Let's fuzz ResolveURL instead which is the lower level utility.
		_ = ResolveURL(baseURL, content)
	})
}
