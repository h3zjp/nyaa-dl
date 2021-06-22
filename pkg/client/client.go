package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/JohnRabin2357/nyaa-dl/pkg/config"
	"github.com/JohnRabin2357/nyaa-dl/pkg/downloader"
	"github.com/JohnRabin2357/nyaa-dl/pkg/nyaa"
)

// Client is the client which downloads .torrent files.
type Client struct {
	config     *config.Config
	nyaa       nyaa.Nyaa
	downloader downloader.Downloader
}

type NewClientInput struct {
	Config     *config.Config
	Nyaa       nyaa.Nyaa
	Downloader downloader.Downloader
}

func New(i *NewClientInput) *Client {
	return &Client{
		config:     i.Config,
		nyaa:       i.Nyaa,
		downloader: i.Downloader,
	}
}

// Run starts downloading for all targets.
func (c *Client) Run(ctx context.Context) error {
	for _, target := range c.config.Targets {
		log.Printf("Running %q", target.Title)
		err := c.runTarget(ctx, target)
		if err != nil {
			return fmt.Errorf("target %q failed: %w", target, err)
		}
	}

	return nil
}

func (c *Client) runTarget(ctx context.Context, target config.Target) error {
	errCnt := 0

	for page := 1; page <= target.MaxPage; page++ {
		torrents, err := c.nyaa.List(ctx, &nyaa.ListInput{
			Domain:   target.Domain,
			Category: target.Category,
			Query:    target.Query,
			Page:     page,
		})
		if err != nil {
			return fmt.Errorf("failed to list torrents: %w", err)
		}

		torrents, err = c.filterTorrents(target, torrents)
		if err != nil {
			return fmt.Errorf("failed to filter torrents: %w", err)
		}
		torrCh := nyaa.GenTorrentsCh(ctx, torrents)

		numWorkers := 2
		workers := make([]<-chan downloadResult, numWorkers)
		for i := 0; i < numWorkers; i++ {
			workers[i] = c.downloadTorrents(ctx, torrCh)
		}

		for res := range c.mergeDownloadResultCh(ctx, workers) {
			if res.Err != nil {
				log.Printf("Failed to download a torrent: %s", res.Err)
				errCnt++

				if errCnt == 5 {
					return fmt.Errorf("failed %d times", errCnt)
				}
				continue
			}
			log.Printf("Downloaded %q", res.Torrent.Title)
		}
	}

	return nil
}

type downloadResult struct {
	Torrent nyaa.Torrent
	Err     error
}

func (c *Client) downloadTorrents(ctx context.Context, torrCh <-chan nyaa.Torrent) <-chan downloadResult {
	resCh := make(chan downloadResult)

	go func() {
		defer close(resCh)
		for tr := range torrCh {
			err := c.downloader.DoWithRetry(ctx, tr)
			if err != nil {
				err = fmt.Errorf("downloading error: %w", err)
			}

			select {
			case <-ctx.Done():
				return
			case resCh <- downloadResult{
				Torrent: tr,
				Err:     err,
			}:
			}
		}
	}()

	return resCh
}

func (c *Client) mergeDownloadResultCh(ctx context.Context, channels []<-chan downloadResult) <-chan downloadResult {
	var wg sync.WaitGroup
	mergedCh := make(chan downloadResult)

	merge := func(ch <-chan downloadResult) {
		defer wg.Done()
		for r := range ch {
			select {
			case <-ctx.Done():
				return
			case mergedCh <- r:
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go merge(ch)
	}
	go func() {
		wg.Wait()
		close(mergedCh)
	}()
	return mergedCh
}

func (c *Client) filterTorrents(target config.Target, torrents []nyaa.Torrent) ([]nyaa.Torrent, error) {
	var filtered []nyaa.Torrent

	for _, tr := range torrents {
		if tr.Downloads < target.RequiredDownloads {
			continue
		}

		done, err := c.downloader.IsDownloaded(tr)
		if err != nil {
			return nil, fmt.Errorf("failed to check whether is a torrent downloaded: %w", err)
		}

		if !done {
			filtered = append(filtered, tr)
		}
	}

	return filtered, nil
}
