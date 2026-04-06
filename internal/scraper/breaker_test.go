package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestService_Scrape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>OK</title></head><body><article>Content here</article></body></html>`))
	}))
	defer srv.Close()

	svc := NewService()
	result, err := svc.Scrape(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "OK" {
		t.Errorf("title = %q, want %q", result.Title, "OK")
	}
}

func TestService_CircuitOpensAfterFailures(t *testing.T) {
	// Server that always returns 500
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	svc := NewService()

	// Trigger 5 failures to open the circuit
	for i := range 5 {
		_, err := svc.Scrape(context.Background(), srv.URL+"/page")
		if err == nil {
			t.Fatalf("attempt %d: expected error, got nil", i+1)
		}
	}

	// 6th attempt should fail fast with circuit breaker open
	_, err := svc.Scrape(context.Background(), srv.URL+"/page")
	if err == nil {
		t.Fatal("expected circuit breaker error")
	}
	if !strings.Contains(err.Error(), "circuit breaker is open") {
		t.Errorf("expected circuit breaker open error, got: %v", err)
	}
}

func TestService_PerDomainBreakers(t *testing.T) {
	// Verify that breakers are keyed per domain by checking that two different
	// domains get independent breaker instances.
	svc := NewService()

	cb1 := svc.getOrCreate("example.com")
	cb2 := svc.getOrCreate("other.com")
	cb3 := svc.getOrCreate("example.com")

	if cb1 == cb2 {
		t.Error("different domains should have different breakers")
	}
	if cb1 != cb3 {
		t.Error("same domain should return same breaker")
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url    string
		want   string
		hasErr bool
	}{
		{"https://example.com/path", "example.com", false},
		{"http://sub.example.com:8080/path", "sub.example.com", false},
		{"not-a-url", "", true},
	}

	for _, tt := range tests {
		got, err := extractDomain(tt.url)
		if tt.hasErr {
			if err == nil {
				t.Errorf("extractDomain(%q) expected error", tt.url)
			}
			continue
		}
		if err != nil {
			t.Errorf("extractDomain(%q) unexpected error: %v", tt.url, err)
			continue
		}
		if got != tt.want {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
