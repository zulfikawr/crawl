# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-01-03

### Added

#### Core Features
- Initial release of Crawl CLI tool
- Hybrid scraping engine with automatic mode detection
- Support for both static HTML and JavaScript-rendered SPAs
- `get` command for content extraction and web scraping
- `media` command for downloading images, videos, and audio files

#### Scraping Engines
- Static engine for fast HTTP-based scraping
- Dynamic engine using headless Chrome for SPAs
- Hybrid engine with intelligent auto-detection
- Batch processing support for multiple URLs
- Metadata extraction utilities

#### CLI Features
- CSS selector-based content extraction
- Multiple output formats: JSON, CSV, HTML, Markdown, plain text
- Custom headers support
- HTTP and SOCKS5 proxy support
- Configurable timeouts and wait times
- User agent customization
- Verbose and quiet logging modes
- JSON-formatted output option

#### Media Downloader
- Concurrent download workers (1-50 configurable)
- Support for images, videos, and audio files
- Stream-based downloads for large files
- Progress tracking with progress bars
- Automatic file type detection

#### Configuration
- YAML configuration file support
- CLI flag-based configuration
- Environment variable support
- Example configuration included

#### Performance & Reliability
- HTTP response caching with TTL
- Automatic retry logic with exponential backoff
- Rate limiting to respect server resources
- Proxy pool management
- Graceful shutdown handling

#### Developer Experience
- Comprehensive test coverage
- Structured logging with zerolog
- Color-coded terminal output
- Progress indicators for long operations
- Request context propagation

### Technical Details

#### Dependencies
- Go 1.25.4
- Cobra for CLI framework
- chromedp for headless browser automation
- goquery for HTML parsing
- zerolog for structured logging
- html-to-markdown for Markdown conversion

#### Supported Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### Known Issues
- Chrome/Chromium must be installed for SPA mode
- Large file downloads may timeout with default settings (increase timeout as needed)

---

[0.1.0]: https://github.com/zulfikawr/crawl/releases/tag/v0.1.0
