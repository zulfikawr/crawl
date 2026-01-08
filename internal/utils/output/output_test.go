package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zulfikawr/crawl/pkg/models"
)

func TestSaveJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.json")

	data := &models.PageData{
		URL:   "http://example.com",
		Title: "Test",
	}

	err := SaveJSON(data, path)
	assert.NoError(t, err)

	// Verify content
	bytes, err := os.ReadFile(path)
	assert.NoError(t, err)

	var loaded models.PageData
	err = json.Unmarshal(bytes, &loaded)
	assert.NoError(t, err)
	assert.Equal(t, data.URL, loaded.URL)
}

func TestSaveMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.md")

	data := &models.PageData{
		URL:  "http://example.com",
		HTML: "<html><body><h1>Test Title</h1><p>Some content</p></body></html>",
	}

	err := SaveMarkdown(data, path)
	assert.NoError(t, err)

	bytes, err := os.ReadFile(path)
	assert.NoError(t, err)
	content := string(bytes)

	assert.Contains(t, content, "# Test Title")
	assert.Contains(t, content, "Some content")
}
