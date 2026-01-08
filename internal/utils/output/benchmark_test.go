package output

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/zulfikawr/crawl/pkg/models"
)

func BenchmarkSaveJSON(b *testing.B) {
	data := &models.PageData{
		URL:     "http://example.com",
		Title:   "Benchmark Title",
		Content: strings.Repeat("Content ", 1000), // ~8KB
		HTML:    strings.Repeat("<div>Content</div>", 1000),
	}
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SaveJSON(data, path)
	}
}

func BenchmarkSaveMarkdown(b *testing.B) {
	data := &models.PageData{
		URL:  "http://example.com",
		HTML: strings.Repeat("<div><p>Paragraph</p><a href='/link'>Link</a></div>", 100),
	}
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.md")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SaveMarkdown(data, path)
	}
}
