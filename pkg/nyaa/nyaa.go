package nyaa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff/v4"
)

// Torrent represents a torrent item of nyaa.
type Torrent struct {
	ID        uint64
	Title     string
	URL       string
	Downloads int
}

// Nyaa provides APIs for nyaa.
type Nyaa interface {
	List(ctx context.Context, input *ListInput) ([]Torrent, error)
}

// ListInput is the input for Nyaa.List().
type ListInput struct {
	// Domain of nyaa.
	Domain string
	// Category to list.
	Category string
	// Query to search.
	Query string
	// Page to list.
	Page int
}

func NewNyaa() Nyaa {
	return &htmlNyaa{
		client: http.DefaultClient,
	}
}

type htmlNyaa struct {
	client *http.Client
}

func (h *htmlNyaa) List(ctx context.Context, input *ListInput) (torrents []Torrent, err error) {
	url := h.buildListURL(input)
	doc, err := h.fetchDocWithRetry(ctx, url)
	if err != nil {
		err = fmt.Errorf("failed to fetch doc of %s: %w", url, err)
		return
	}

	var parseErr error
	doc.Find("table.torrent-list > tbody > tr.success").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		tr := Torrent{}

		// Title
		tr.Title = s.Find("td:nth-child(2) > a:last-child").Text()
		if tr.Title == "" {
			parseErr = errors.New("failed to get a torrent title, maybe nyaa html structure has changed")
			return false
		}

		// Downloads
		dlStr := s.Find("td:last-child").Text()
		if dlStr == "" {
			parseErr = errors.New("failed to get a num of downloads of torrent, maybe nyaa html structure has changed")
			return false
		}
		dlNum, err := strconv.Atoi(dlStr)
		if err != nil {
			parseErr = fmt.Errorf("failed to covert to num from %q: %w", dlStr, err)
			return false
		}
		tr.Downloads = dlNum

		// URL
		href, ok := s.Find("td:nth-child(3) > a:first-child").Attr("href")
		if !ok {
			parseErr = errors.New("failed to get a URL of torrent, maybe nyaa html structure has changed")
			return false
		}
		tr.URL = fmt.Sprintf("https://%s/%s", input.Domain, strings.TrimPrefix(href, "/"))

		// ID
		id, err := extractID(tr.URL)
		if err != nil {
			parseErr = fmt.Errorf("failed to get ID of torrent: %w", err)
			return false
		}
		tr.ID = id

		// Done
		torrents = append(torrents, tr)
		return true
	})
	if parseErr != nil {
		err = fmt.Errorf("failed to parse a html of nyaa: %w", err)
	}

	return
}

// extractID extracts a torrent ID of URL.
// URL must be like "https://nyaa.example.com/view/12345.torrent".
func extractID(url string) (uint64, error) {
	strs := strings.Split(url, "/")
	if len(strs) == 1 {
		return 0, errors.New("invalid URL")
	}

	idStr := strings.TrimSuffix(strs[len(strs)-1], ".torrent")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ID as uint: %w", err)
	}

	return id, nil
}

func (h *htmlNyaa) fetchDocWithRetry(ctx context.Context, url string) (*goquery.Document, error) {
	var doc *goquery.Document

	err := backoff.Retry(func() error {
		d, err := h.fetchDoc(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to fetch a goquery document: %w", err)
		}
		doc = d
		return nil
	}, backoff.WithContext(backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5), ctx))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch a goquery document with retry: %w", err)
	}

	return doc, nil
}

func (h *htmlNyaa) fetchDoc(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build a http request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http status code is not 200, got %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery failed to read a response body: %w", err)
	}

	return doc, nil
}

func (h *htmlNyaa) buildListURL(input *ListInput) string {
	params := url.Values{}

	params.Set("f", "2") // Trusted only

	params.Set("c", input.Category)
	params.Set("p", fmt.Sprintf("%d", input.Page))

	if input.Query != "" {
		params.Set("q", input.Query)
	}

	return fmt.Sprintf("https://%s/?%s", input.Domain, params.Encode())
}
