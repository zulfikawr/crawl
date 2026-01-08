// internal/cli/get.go
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/crawl/internal/config"
	"github.com/zulfikawr/crawl/internal/engine"
	"github.com/zulfikawr/crawl/internal/ui"
	headersutil "github.com/zulfikawr/crawl/internal/utils/headers"
	outpututil "github.com/zulfikawr/crawl/internal/utils/output"
	urlutil "github.com/zulfikawr/crawl/internal/utils/url"
	"github.com/zulfikawr/crawl/pkg/models"
)

var (
	mode           string
	selector       string
	output         string
	headers        []string
	fields         string
	getWaitSeconds int
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <url>",
	Short: "Retrieve text or data from a URL",
	Long: `Intelligently switches between the "Fast Engine" (HTTP/Static) and 
"Deep Engine" (Headless/SPA) to get raw HTML or structured data.

The scraper will auto-detect whether to use static or SPA mode, or you can 
force a specific mode using the --mode flag.`,
	Example: `  # Basic scrape (auto-detects static vs SPA)
  crawl get https://example.com

  # Force static mode for speed
  crawl get https://example.com --mode=static

  # Extract specific content with CSS selector
  crawl get https://example.com --selector=".price-tag"

  # Save output to JSON file
  crawl get https://example.com --output=data.json

  # Add custom headers
  crawl get https://example.com -H "Authorization: Bearer token"`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&mode, "mode", "m", "auto", "Force engine mode: auto, static, or spa")
	getCmd.Flags().StringVarP(&selector, "selector", "s", "body", "CSS selector to extract (e.g., .price, #content)")
	getCmd.Flags().StringVarP(&output, "output", "o", "", "File path to save output (supports .json, .txt, .html, .csv, .md)")
	getCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Custom headers (e.g., -H \"User-Agent: Bot\")")
	getCmd.Flags().IntVar(&getWaitSeconds, "wait", 0, "Seconds to wait after page loads before scraping (for dynamic content)")
	getCmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields for CSV export (e.g., name=.name,price=.price)")
}

func runGet(cmd *cobra.Command, args []string) error {
	url := args[0]

	// Validate URL
	if err := urlutil.ValidateURL(url); err != nil {
		cmd.SilenceUsage = false // Show usage for validation errors
		return ui.NewHintedError(err, "Please provide a valid URL including scheme (e.g., https://google.com)")
	}

	// Parse mode
	scraperMode := models.ModeAuto
	switch strings.ToLower(mode) {
	case "auto":
		scraperMode = models.ModeAuto
	case "static":
		scraperMode = models.ModeStatic
	case "spa":
		scraperMode = models.ModeSPA
	default:
		return fmt.Errorf("invalid mode: %s (must be auto, static, or spa)", mode)
	}

	// Warn if using default broad selector (now using styled warn)
	if selector == "body" {
		fmt.Println(ui.StyledWarn("Using default 'body' selector extracts entire page. Use --selector for specific content."))
		fmt.Println()
	}

	// Parse custom headers
	headerMap := headersutil.ParseHeaders(headers)

	// Add user agent if configured globally
	if userAgent != "" && headerMap["User-Agent"] == "" {
		headerMap["User-Agent"] = userAgent
	}

	// Parse fields
	fieldsMap := make(map[string]string)
	if fields != "" {
		pairs := strings.Split(fields, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				fieldsMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Build request options
	opts := models.RequestOptions{
		URL:         url,
		Mode:        scraperMode,
		Selector:    selector,
		Fields:      fieldsMap,
		Headers:     headerMap,
		Timeout:     30 * time.Second,
		Proxy:       proxy, // Global proxy flag
		WaitSeconds: getWaitSeconds,
	}

	// Parse timeout from global flag
	if timeout != "" {
		duration, err := time.ParseDuration(timeout)
		if err != nil {
			log.Warn().Str("timeout", timeout).Msg("Invalid timeout format, using default 30s")
		} else {
			opts.Timeout = duration
		}
	}

	// Select scraper based on requested mode
	var scraper engine.Scraper

	// Get app from command context
	appCtx := GetAppFromCmd(cmd)
	if appCtx == nil {
		return fmt.Errorf("application not initialized")
	}

	// Default: application-level scraper (hybrid)
	scraper = appCtx.Scraper

	switch scraperMode {
	case models.ModeStatic:
		if appCtx.StaticScraper != nil {
			scraper = appCtx.StaticScraper
		}
	case models.ModeSPA:
		// Ensure browser pool exists before using the dynamic scraper
		if appCtx.DynamicScraper == nil {
			return fmt.Errorf("dynamic scraper is unavailable")
		}

		ctx, cancel := context.WithTimeout(context.Background(), appCtx.Config.HTTPTimeout*2)
		defer cancel()
		if appCtx.BrowserPool == nil {
			// Optimization: For single URL fetches, we only need 1 browser context.
			// If the user hasn't overridden the default pool size, reduce it to 1.
			if appCtx.Config.BrowserPoolSize == config.DefaultBrowserPoolSize {
				log.Debug().Msg("Optimizing browser pool size to 1 for single request")
				appCtx.Config.BrowserPoolSize = 1
			}

			if err := appCtx.EnsureBrowserPool(ctx); err != nil {
				// We don't return error here yet, but we'll get it during Fetch
				// We use StyledWarn instead of log.Warn for better visibility
				fmt.Println(ui.StyledWarn("Could not initialize browser pool. Dynamic scraping may fail."))
				fmt.Println(ui.StyledHint("Check if Chrome/Chromium is installed or set CHROME_PATH."))
			}
		}
		scraper = appCtx.DynamicScraper
	}

	// Fetch data with spinner
	spinner := ui.NewSpinner(fmt.Sprintf("Fetching %s...", url))
	spinner.Start()

	pageData, err := scraper.Fetch(opts)

	spinner.Stop()

	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return ui.NewHintedError(err, "Google Chrome is required for SPA mode. Install Chrome or use --mode=static for faster, static-only scraping.")
		}
		return fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Handle output
	if output != "" {
		return saveOutput(pageData, output)
	}

	// Print to stdout
	return printOutput(pageData)
}

func saveOutput(data *models.PageData, pathStr string) error {
	// Normalize extension checks to be case-insensitive
	path := strings.ToLower(pathStr)

	switch {
	case strings.HasSuffix(path, ".json"):
		if err := outpututil.SaveJSON(data, pathStr); err != nil {
			return fmt.Errorf("failed to save JSON: %w", err)
		}
	case strings.HasSuffix(path, ".html"):
		cleaned, err := outpututil.CleanHTML(data.HTML)
		if err != nil {
			return fmt.Errorf("failed to clean HTML: %w", err)
		}
		if err := os.WriteFile(pathStr, []byte(cleaned), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	case strings.HasSuffix(path, ".txt"):
		if err := os.WriteFile(pathStr, []byte(data.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	case strings.HasSuffix(path, ".csv"):
		if err := outpututil.SaveCSV(data, pathStr); err != nil {
			return fmt.Errorf("failed to save CSV: %w", err)
		}
	case strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown"):
		if err := outpututil.SaveMarkdown(data, pathStr); err != nil {
			return fmt.Errorf("failed to save Markdown: %w", err)
		}
	default:
		// Fallback to JSON for unknown extensions
		if err := outpututil.SaveJSON(data, pathStr); err != nil {
			return fmt.Errorf("failed to save JSON: %w", err)
		}
	}

	// Print metadata summary for saved outputs (single call)
	printMetadataSummary(data)

	// Make clickable link when possible using OSC 8 terminal hyperlink
	link := terminalHyperlink(filepath.Base(pathStr), pathStr)
	fmt.Println(ui.StyledSuccess("Saved to " + link))
	return nil
}

// terminalHyperlink returns an OSC 8 hyperlink if supported, falling back to plain path
func terminalHyperlink(label, target string) string {
	abs, err := filepath.Abs(target)
	if err != nil {
		abs = target
	}
	// OSC 8 hyperlink: ESC ] 8 ;; url BEL label ESC ] 8 ;; BEL
	// Use file:// scheme for local files
	return fmt.Sprintf("\x1b]8;;file://%s\x1b\\%s\x1b]8;;\x1b\\", abs, label)
}

// printMetadataSummary prints key metadata fields from PageData using colors and aligns columns
func printMetadataSummary(data *models.PageData) {
	fmt.Println(ui.Heading("Page Metadata"))

	labelStyled := func(s string) string { return ui.ColorBold + s + ui.ColorReset }
	valStyled := func(s string) string { return s }

	rows := []struct {
		Label string
		Value string
	}{
		{"Url", data.URL},
		{"Status", fmt.Sprintf("%d", data.StatusCode)},
		{"Title", data.Title},
		{"Time", fmt.Sprintf("%dms", data.ResponseTime)},
		{"Links", fmt.Sprintf("%d", len(data.Links))},
		{"Images", fmt.Sprintf("%d", len(data.Images))},
		{"Scripts", fmt.Sprintf("%d", len(data.Scripts))},
	}

	var maxLen int
	for _, r := range rows {
		if len(r.Label) > maxLen {
			maxLen = len(r.Label)
		}
	}

	for _, r := range rows {
		pad := strings.Repeat(" ", maxLen-len(r.Label))
		fmt.Printf("  %s%s : %s\n", labelStyled(r.Label), pad, valStyled(r.Value))
	}
}

func printOutput(data *models.PageData) error {
	// If JSON output is requested
	if jsonOutput {
		// Create a copy to avoid modifying the original data
		exportData := *data
		exportData.HTML = "" // Remove HTML from JSON export
		urlutil.ResolveRelativeLinks(&exportData)
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(exportData)
	}

	// If selector was used, print just the content
	if selector != "" && selector != "body" {
		fmt.Println(ui.Heading("Extracted Content"))
		fmt.Println(data.Content)
		return nil
	}

	// Otherwise, print a summary with colors
	printMetadataSummary(data)

	// Print content preview
	contentPreview := data.Content
	if len(contentPreview) > 500 {
		contentPreview = contentPreview[:500] + "..."
	}

	fmt.Println()
	fmt.Println(ui.Heading("Content Preview"))
	fmt.Printf("  %s\n", contentPreview)

	// Helpful hint for saving to a file
	fmt.Println()
	fmt.Println(ui.StyledHint("Use --output=<file> to save to .json, .txt, .html, .csv, or .md"))

	return nil
}
