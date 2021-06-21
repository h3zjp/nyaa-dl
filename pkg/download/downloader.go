package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/JohnRabin2357/nyaa-dl/pkg/nyaa"
	"github.com/cenkalti/backoff/v4"
)

type Downloader interface {
	Do(ctx context.Context, tr nyaa.Torrent) error
	DoWithRetry(ctx context.Context, tr nyaa.Torrent) error
	IsDownloaded(tr nyaa.Torrent) (bool, error)
}

func NewDownloader(outputPath string) Downloader {
	return &downloader{
		outputPath: outputPath,
		client:     http.DefaultClient,
	}
}

type downloader struct {
	outputPath string
	client     *http.Client
}

func (d *downloader) DoWithRetry(ctx context.Context, tr nyaa.Torrent) error {
	maxRetries := uint64(5)

	err := backoff.Retry(func() error {
		if err := d.Do(ctx, tr); err != nil {
			return fmt.Errorf("download error: %w", err)
		}
		return nil
	}, backoff.WithContext(backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxRetries), ctx))

	if err != nil {
		return fmt.Errorf("failed %d times, and the last err is: %w", maxRetries, err)
	}
	return nil
}

func (d *downloader) Do(ctx context.Context, tr nyaa.Torrent) error {
	req, err := http.NewRequestWithContext(ctx, "GET", tr.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to build a http request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("http status code is not 200, got %d", resp.StatusCode)
	}

	file, err := os.OpenFile(d.path(tr), os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		return fmt.Errorf("failed to open a file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		// Remove crashed file
		errRemove := os.Remove(file.Name())
		if errRemove != nil {
			return fmt.Errorf("file copying error and couldn't remove a crashed file (%s): %w", file.Name(), errRemove)
		}

		return fmt.Errorf("file copying error: %w", err)
	}

	return nil
}

func (d *downloader) IsDownloaded(tr nyaa.Torrent) (bool, error) {
	_, err := os.Stat(d.outputPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(d.outputPath, 0775)
		if err != nil {
			return false, fmt.Errorf("failed to make output dir: %w", err)
		}
	}
	if err != nil {
		return false, fmt.Errorf("failed to check whether does output dir exist: %w", err)
	}

	_, err = os.Stat(d.path(tr))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check whether does a file exist: %w", err)
	}

	return true, nil
}

// path returns the path to save a torrent.
func (d *downloader) path(tr nyaa.Torrent) string {
	return filepath.Join(d.outputPath, fmt.Sprintf("%d.torrent", tr.ID))
}
