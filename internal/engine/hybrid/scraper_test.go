package hybrid

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/zulfikawr/crawl/pkg/models"
)

func TestExecuteScripts(t *testing.T) {
	html := `
	<html>
		<body>
			<script>
				var config = {
					apiKey: "12345",
					version: 1
				};
				var simpleString = "hello world";
			</script>
		</body>
	</html>
	`

	data := &models.PageData{
		URL:      "http://example.com",
		HTML:     html,
		Metadata: make(map[string]string),
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	assert.NoError(t, err)

	executeScripts(data, doc)

	// Verify JS variables were extracted
	assert.Contains(t, data.Metadata, "js:simpleString")
	assert.Equal(t, "hello world", data.Metadata["js:simpleString"])

	assert.Contains(t, data.Metadata, "js:config")
	// The map string representation might vary in order, so check substring
	assert.Contains(t, data.Metadata["js:config"], "apiKey:12345")
}
