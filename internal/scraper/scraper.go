package scraper

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Result holds the extracted content from a scraped URL.
type Result struct {
	Title   string
	Content string
}

var userAgents = []string{
	"Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
}

var whitespaceRe = regexp.MustCompile(`\s+`)

// Scrape fetches and extracts clean text content from the given URL.
func Scrape(ctx context.Context, targetURL string) (Result, error) {
	var result Result
	var scrapeErr error

	c := colly.NewCollector(
		colly.UserAgent(randomUA()),
	)
	c.SetRequestTimeout(30 * time.Second)

	c.OnHTML("title", func(e *colly.HTMLElement) {
		if result.Title == "" {
			result.Title = strings.TrimSpace(e.Text)
		}
	})

	c.OnHTML("article, main, [role='main']", func(e *colly.HTMLElement) {
		if result.Content != "" {
			return // already found content from a higher-priority element
		}
		e.DOM.Find("script, style, nav, footer, header").Remove()
		text := cleanText(e.Text)
		if text != "" {
			result.Content = text
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		if result.Content != "" {
			return // prefer article/main over body
		}
		e.DOM.Find("script, style, nav, footer, header").Remove()
		text := cleanText(e.Text)
		if text != "" {
			result.Content = text
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		scrapeErr = err
	})

	// Use a channel to respect context cancellation.
	// Note: colly does not support mid-request cancellation, so the goroutine
	// may outlive the context. We still wait for it to finish to avoid leaking.
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := c.Visit(targetURL); err != nil {
			scrapeErr = err
		}
	}()

	select {
	case <-ctx.Done():
		// Wait for the goroutine to finish to prevent a leak.
		// The collector's request timeout (30s) bounds this wait.
		<-done
		return Result{}, fmt.Errorf("scrape cancelled: %w", ctx.Err())
	case <-done:
	}

	if scrapeErr != nil {
		return Result{}, fmt.Errorf("scraping %s: %w", targetURL, scrapeErr)
	}

	if result.Content == "" {
		return Result{}, fmt.Errorf("no content extracted from %s", targetURL)
	}

	return result, nil
}

func randomUA() string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents))))
	if err != nil {
		return userAgents[0]
	}
	return userAgents[n.Int64()]
}

func cleanText(s string) string {
	s = whitespaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
