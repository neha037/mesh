package scraper

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"
)

// Service wraps Scrape with per-domain circuit breakers.
type Service struct {
	mu       sync.Mutex
	breakers map[string]*gobreaker.CircuitBreaker[Result]
}

// NewService creates a new scraper Service with circuit breaker protection.
func NewService() *Service {
	return &Service{
		breakers: make(map[string]*gobreaker.CircuitBreaker[Result]),
	}
}

// Scrape fetches content from the URL, protected by a per-domain circuit breaker.
func (s *Service) Scrape(ctx context.Context, targetURL string) (Result, error) {
	domain, err := extractDomain(targetURL)
	if err != nil {
		return Result{}, err
	}

	cb := s.getOrCreate(domain)

	result, err := cb.Execute(func() (Result, error) {
		return Scrape(ctx, targetURL)
	})
	if err != nil {
		return Result{}, fmt.Errorf("scrape %s (breaker: %s): %w", targetURL, cb.State().String(), err)
	}
	return result, nil
}

func (s *Service) getOrCreate(domain string) *gobreaker.CircuitBreaker[Result] {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cb, ok := s.breakers[domain]; ok {
		return cb
	}

	cb := gobreaker.NewCircuitBreaker[Result](gobreaker.Settings{
		Name:        "scraper-" + domain,
		MaxRequests: 1,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})

	s.breakers[domain] = cb
	return cb
}

func extractDomain(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL %q: %w", rawURL, err)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("URL %q has no host", rawURL)
	}
	return parsed.Hostname(), nil
}
