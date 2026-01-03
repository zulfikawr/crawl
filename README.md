# Crawl

A fast and cross-platform CLI tool for scraping websites and downloading media files. Crawl intelligently handles both static HTML pages and JavaScript-rendered Single Page Applications (SPAs) with automatic detection and hybrid mode support.

## Features

- **Hybrid Scraping Engine**: Automatically detects and switches between static HTTP requests and headless Chrome for SPAs
- **Smart Content Extraction**: Extract specific content using CSS selectors
- **Media Downloader**: Download images, videos, and audio files with concurrent workers
- **Multiple Output Formats**: Save results as JSON, CSV, HTML, Markdown, or plain text
- **Flexible Configuration**: Configure via CLI flags, YAML config files, or environment variables
- **Proxy Support**: HTTP and SOCKS5 proxy support for anonymity
- **Custom Headers**: Add authentication tokens, user agents, and custom headers
- **Rate Limiting**: Built-in rate limiting to respect server resources
- **Retry Logic**: Automatic retry with exponential backoff for failed requests
- **Progress Tracking**: Real-time progress bars for long-running operations
- **Caching**: Intelligent caching to avoid redundant requests

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/zulfikawr/crawl/releases).

### Build from Source

Requires Go 1.23 or later:

```bash
git clone https://github.com/zulfikawr/crawl.git
cd crawl
go build -o crawl ./cmd/crawl
```

### Install with Go

```bash
go install github.com/zulfikawr/crawl/cmd/crawl@latest
```

## Quick Start

### Basic Web Scraping

```bash
# Scrape a web page (auto-detects static vs SPA)
crawl get https://example.com

# Extract specific content with CSS selector
crawl get https://example.com --selector=".article-title"

# Save output to a file
crawl get https://example.com --output=data.json

# Force static mode for faster scraping
crawl get https://example.com --mode=static

# Force SPA mode for JavaScript-heavy sites
crawl get https://news.ycombinator.com --mode=spa
```

### Media Downloads

```bash
# Download all images from a page
crawl media https://example.com --type=image

# Download videos with custom concurrency
crawl media https://example.com --type=video --concurrency=10

# Download all media types to a specific directory
crawl media https://example.com --type=all --output=./my-downloads

# Handle JavaScript-rendered media galleries
crawl media https://spa-gallery.com --mode=spa --type=image
```

### Advanced Usage

```bash
# Add custom headers (e.g., authentication)
crawl get https://api.example.com -H "Authorization: Bearer token123"

# Use a proxy
crawl get https://example.com --proxy=http://proxy.example.com:8080

# Set custom timeout
crawl get https://slow-site.com --timeout=60s

# Wait for dynamic content to load
crawl get https://dynamic-site.com --wait=5

# Extract CSV data with custom fields
crawl get https://example.com --output=data.csv --fields="title=.title,price=.price"

# Enable verbose logging for debugging
crawl get https://example.com --verbose
```

## Commands

### `crawl get`

Retrieve and extract content from a URL.

**Usage:**
```bash
crawl get <url> [flags]
```

**Flags:**
- `-m, --mode string` - Force engine mode: auto, static, or spa (default "auto")
- `-s, --selector string` - CSS selector to extract (default "body")
- `-o, --output string` - File path to save output (supports .json, .txt, .html, .csv, .md)
- `-H, --header stringArray` - Custom headers (can be specified multiple times)
- `--wait int` - Seconds to wait after page loads before scraping
- `--fields string` - Comma-separated fields for CSV export (e.g., "name=.name,price=.price")

### `crawl media`

Download media files from a URL.

**Usage:**
```bash
crawl media <url> [flags]
```

**Flags:**
- `-t, --type string` - Media type: image, video, audio, or all (default "all")
- `-o, --output string` - Directory to save files (default "./downloads")
- `-c, --concurrency int` - Number of concurrent workers (1-50) (default 5)
- `-m, --mode string` - Scraper mode: auto, static, or spa (default "auto")
- `-H, --header stringArray` - Custom headers
- `--wait int` - Seconds to wait after page loads

## Global Flags

- `--config string` - Path to YAML configuration file
- `--proxy string` - HTTP/SOCKS5 proxy URL
- `--timeout string` - Hard timeout for requests (default "30s")
- `--user-agent string` - Custom user agent string
- `-v, --verbose` - Enable debug logging
- `-q, --quiet` - Suppress all output except errors
- `--json` - Output in JSON format only
- `-h, --help` - Show help
- `--version` - Show version

## Configuration

### Configuration File

Create a `crawl.yaml` file in your working directory or specify with `--config`:

```yaml
# Logging
log_level: info
json_log: false

# HTTP Settings
http_timeout: 30s
user_agent: "Crawl/1.0 (https://github.com/zulfikawr/crawl)"

# Cache Settings
cache_ttl: 5m
cache_max_size_bytes: 104857600  # 100MB

# Browser Settings (for SPA mode)
browser_pool_size: 3
browser_headless: true
chrome_path: ""  # Auto-detect if empty
```

See [internal/config/examples/crawl.yaml](internal/config/examples/crawl.yaml) for a complete example.

### Environment Variables

Configuration values can also be set via environment variables:

```bash
export CRAWL_LOG_LEVEL=debug
export CRAWL_HTTP_TIMEOUT=60s
export CRAWL_PROXY=http://proxy.example.com:8080
```

## Architecture

Crawl uses a modular architecture with three scraping engines:

- **Static Engine**: Fast HTTP-based scraping for traditional HTML pages
- **Dynamic Engine**: Headless Chrome for JavaScript-rendered SPAs
- **Hybrid Engine**: Automatically detects and switches between modes

### Project Structure

```
crawl/
├── cmd/crawl/          # CLI entry point
├── internal/
│   ├── app/            # Application initialization
│   ├── cache/          # HTTP caching layer
│   ├── cli/            # Command definitions
│   ├── config/         # Configuration management
│   ├── downloader/     # Media download logic
│   ├── engine/         # Scraping engines
│   │   ├── static/     # Static HTML scraper
│   │   ├── dynamic/    # Headless browser scraper
│   │   ├── hybrid/     # Auto-detection logic
│   │   ├── batch/      # Batch processing
│   │   └── metadata/   # Metadata extraction
│   ├── proxy/          # Proxy pool management
│   ├── ratelimit/      # Rate limiting
│   ├── retry/          # Retry logic
│   ├── ui/             # Terminal UI components
│   └── utils/          # Utility functions
└── pkg/models/         # Shared data models
```

## Requirements

- **Go**: 1.23 or later (for building from source)
- **Chrome/Chromium**: Required for SPA mode (automatically detected or specify with `chrome_path`)
- **Operating Systems**: Linux, macOS, Windows

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/engine/static/
```

### Building

```bash
# Build for current platform
go build -o crawl ./cmd/crawl

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o crawl-linux-amd64 ./cmd/crawl
GOOS=darwin GOARCH=amd64 go build -o crawl-darwin-amd64 ./cmd/crawl
GOOS=windows GOARCH=amd64 go build -o crawl-windows-amd64.exe ./cmd/crawl
```

## Use Cases

- **Price Monitoring**: Extract product prices and availability
- **Content Aggregation**: Collect articles, blog posts, or news
- **Data Mining**: Extract structured data from websites
- **Media Archival**: Download images and videos from galleries
- **Web Testing**: Validate page content and structure
- **Research**: Collect data for analysis and research projects

## Performance Tips

1. **Use Static Mode**: When possible, use `--mode=static` for 10-100x faster scraping
2. **Increase Concurrency**: For media downloads, increase `--concurrency` based on your bandwidth
3. **Enable Caching**: Configure cache settings to avoid redundant requests
4. **Specific Selectors**: Use precise CSS selectors instead of `body` to reduce data processing
5. **Adjust Timeouts**: Set appropriate timeouts based on site speed

## Troubleshooting

### Chrome/Chromium Not Found

If SPA mode fails with a Chrome error:

```bash
# Specify Chrome path in config
chrome_path: "/usr/bin/chromium-browser"

# Or use environment variable
export CRAWL_CHROME_PATH=/usr/bin/google-chrome
```

### Rate Limited by Website

```bash
# Use proxy rotation
crawl get https://example.com --proxy=http://proxy1.example.com:8080

# Add delays between requests (via config)
rate_limit_requests: 10
rate_limit_duration: 1s
```

### JavaScript Content Not Loading

```bash
# Increase wait time
crawl get https://example.com --wait=10

# Force SPA mode
crawl get https://example.com --mode=spa
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [chromedp](https://github.com/chromedp/chromedp) for headless browsing
- Uses [goquery](https://github.com/PuerkitoBio/goquery) for HTML parsing
- Powered by [zerolog](https://github.com/rs/zerolog) for structured logging

## Support

- **Issues**: [GitHub Issues](https://github.com/zulfikawr/crawl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/zulfikawr/crawl/discussions)