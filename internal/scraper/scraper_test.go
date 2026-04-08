package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestScrape(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantTitle   string
		wantContent string
		wantErr     bool
	}{
		{
			name: "extracts title and article content",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(`<html><head><title>Test Page</title></head>
					<body><article>Hello world, this is content.</article></body></html>`))
			},
			wantTitle:   "Test Page",
			wantContent: "Hello world, this is content.",
		},
		{
			name: "falls back to body when no article",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(`<html><head><title>Body Page</title></head>
					<body><p>Body content here.</p></body></html>`))
			},
			wantTitle:   "Body Page",
			wantContent: "Body content here.",
		},
		{
			name: "strips script and style tags",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(`<html><head><title>Clean</title></head>
					<body><article><script>alert('xss')</script>
					<style>.foo{}</style>Real content.</article></body></html>`))
			},
			wantTitle:   "Clean",
			wantContent: "Real content.",
		},
		{
			name: "returns error on 404",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			result, err := Scrape(context.Background(), srv.URL)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", result.Title, tt.wantTitle)
			}
			if !strings.Contains(result.Content, tt.wantContent) {
				t.Errorf("content = %q, want to contain %q", result.Content, tt.wantContent)
			}
		})
	}
}

func TestScrape_RespectsRobotsTxt(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /blocked\n"))
	})
	mux.HandleFunc("/blocked", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Blocked</title></head><body><article>Secret content</article></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, err := Scrape(context.Background(), srv.URL+"/blocked")
	if err == nil {
		t.Fatal("expected error when scraping robots.txt-blocked URL, got nil")
	}
}

func TestScrape_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = w.Write([]byte(`<html><body>slow</body></html>`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := Scrape(ctx, srv.URL)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
