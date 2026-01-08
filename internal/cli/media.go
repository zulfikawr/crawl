// internal/cli/media.go
package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/crawl/internal/downloader"
	"github.com/zulfikawr/crawl/internal/engine"
	"github.com/zulfikawr/crawl/internal/ui"
	headersutil "github.com/zulfikawr/crawl/internal/utils/headers"
	urlutil "github.com/zulfikawr/crawl/internal/utils/url"
	"github.com/zulfikawr/crawl/pkg/models"
)

var (
	mediaType        string
	concurrency      int
	outputDir        string
	mediaWaitSeconds int
)

// mediaCmd represents the media command
var mediaCmd = &cobra.Command{
	Use:   "media <url>",
	Short: "Download media files (images, videos, audio) from a URL",
	Long: `Extracts and downloads media files from a web page using concurrent workers.

The media command can:
  - Extract images, videos, or audio files from any web page
  - Download multiple files concurrently using a worker pool
  - Handle both static HTML and JavaScript-rendered SPAs
  - Stream large files to disk without loading into RAM

Perfect for downloading galleries, video collections, or scraping media-heavy sites.`,
	Example: `  # Download all images from a page
  crawl media https://example.com --type=image

  # Download videos with 10 concurrent workers
  crawl media https://example.com/videos --type=video --concurrency=10

  # Download all media types to a specific directory
  crawl media https://example.com --type=all --output=./downloads

  # Download from a SPA that requires JavaScript
  crawl media https://spa-site.com --mode=spa --type=video`,
	Args: cobra.ExactArgs(1),
	RunE: runMedia,
}

func init() {
	rootCmd.AddCommand(mediaCmd)

	mediaCmd.Flags().StringVarP(&mediaType, "type", "t", "all", "Media type to download: image, video, audio, or all")
	mediaCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 5, "Number of concurrent download workers (1-50)")
	mediaCmd.Flags().StringVarP(&outputDir, "output", "o", "./downloads", "Directory to save downloaded files")
	mediaCmd.Flags().StringVarP(&mode, "mode", "m", "auto", "Scraper mode: auto, static, or spa")
	mediaCmd.Flags().IntVar(&mediaWaitSeconds, "wait", 0, "Seconds to wait after page loads before scraping (static and SPA)")
	mediaCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Custom headers")

}

func runMedia(cmd *cobra.Command, args []string) error {
	pageURL := args[0]

	// Validate URL
	if err := urlutil.ValidateURL(pageURL); err != nil {
		cmd.SilenceUsage = false
		return ui.NewHintedError(err, "Please provide a valid URL including scheme (e.g., https://google.com)")
	}

	// Validate media type
	var mediaTypeEnum downloader.MediaType
	switch strings.ToLower(mediaType) {
	case "image", "img":
		mediaTypeEnum = downloader.MediaTypeImage
	case "video", "vid":
		mediaTypeEnum = downloader.MediaTypeVideo
	case "audio":
		mediaTypeEnum = downloader.MediaTypeAudio
	case "all":
		mediaTypeEnum = downloader.MediaTypeAll
	default:
		return fmt.Errorf("invalid media type: %s (must be image, video, audio, or all)", mediaType)
	}

	// Validate concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > 50 {
		concurrency = 50
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
	}

	// Parse custom headers
	headerMap := headersutil.ParseHeaders(headers)

	// Create scraper to fetch the page
	var scraper engine.Scraper

	// Get app from command context
	appCtx := GetAppFromCmd(cmd)
	if appCtx == nil {
		return fmt.Errorf("application not initialized")
	}

	// Use the scraper from the app
	scraper = appCtx.Scraper

	if scraperMode == models.ModeSPA {
		ctx, cancel := context.WithTimeout(context.Background(), appCtx.Config.HTTPTimeout*2)
		defer cancel()
		if appCtx.BrowserPool == nil {
			if err := appCtx.EnsureBrowserPool(ctx); err != nil {
				fmt.Println(ui.StyledWarn("Could not initialize browser pool. Dynamic scraping may fail."))
				fmt.Println(ui.StyledHint("Check if Chrome/Chromium is installed or set CHROME_PATH."))
			}
		}
	}

	// Fetch the page with spinner
	spinner := ui.NewSpinner(fmt.Sprintf("Fetching %s...", pageURL))
	spinner.Start()

	opts := models.RequestOptions{
		URL:         pageURL,
		Mode:        scraperMode,
		Headers:     headerMap,
		Timeout:     30 * time.Second,
		WaitSeconds: mediaWaitSeconds,
	}

	pageData, err := scraper.Fetch(opts)
	spinner.Stop()

	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return ui.NewHintedError(err, "Google Chrome is required for SPA mode. Install Chrome or use --mode=static.")
		}
		return fmt.Errorf("failed to fetch page: %w", err)
	}

	// Extract media URLs from the HTML
	mediaURLs, err := downloader.ExtractMedia(pageData.HTML, pageURL, mediaTypeEnum)
	if err != nil {
		return fmt.Errorf("failed to extract media: %w", err)
	}

	if len(mediaURLs) == 0 {
		fmt.Println(ui.StyledError("No media files found."))
		fmt.Println(ui.StyledHint("Try using --mode=spa for JavaScript-heavy sites"))
		return nil
	}

	fmt.Println(ui.Heading("Media Discovery"))
	fmt.Printf("  %s %s\n", ui.Bold("Found"), fmt.Sprintf("%d media file(s).", len(mediaURLs)))
	fmt.Println()

	// Create output directory
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}

	// Create worker pool
	pool := downloader.NewWorkerPool(concurrency, 60*time.Second, "Crawl/1.0")

	// Start downloads
	fmt.Println(ui.Heading("Downloading"))
	fmt.Printf("  %s %s\n", ui.Bold("Using"), fmt.Sprintf("%d concurrent workers", concurrency))
	fmt.Println()
	ctx := context.Background()

	downloadOpts := downloader.DownloadOptions{
		OutputDir: absOutputDir,
		Headers:   headerMap,
	}

	// Reduce console logging during the download phase
	prevLevel := zerolog.GlobalLevel()
	if !verbose && !jsonOutput {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	results := pool.DownloadBatch(ctx, mediaURLs, downloadOpts)
	// Restore previous log level
	zerolog.SetGlobalLevel(prevLevel)

	// Print results
	successCount := 0
	failCount := 0
	totalSize := int64(0)
	totalDuration := time.Duration(0)

	for _, result := range results {
		if result.Success {
			successCount++
			totalSize += result.Size
			totalDuration += result.Duration
		} else {
			failCount++
		}
	}

	// Compute average duration
	avgDuration := time.Duration(0)
	if successCount > 0 {
		avgDuration = totalDuration / time.Duration(successCount)
	}

	fmt.Println()
	fmt.Println(ui.Heading("Summary"))
	fmt.Printf("  %s %s\n", ui.Bold("Total    :"), fmt.Sprintf("%d files", len(results)))
	fmt.Printf("  %s %s\n", ui.Bold("Success  :"), ui.Success(fmt.Sprintf("%d", successCount)))
	fmt.Printf("  %s %s\n", ui.Bold("Failed   :"), ui.Error(fmt.Sprintf("%d", failCount)))
	fmt.Printf("  %s %s\n", ui.Bold("Size     :"), ui.FormatBytes(totalSize))
	if successCount > 0 {
		fmt.Printf("  %s %s\n", ui.Bold("Avg Time :"), avgDuration.Round(time.Millisecond).String())
	}
	fmt.Printf("  %s %s\n", ui.Bold("Output   :"), absOutputDir)

	if failCount > 0 {
		cmd.SilenceUsage = true
		return fmt.Errorf("%d download(s) failed", failCount)
	}

	return nil
}
