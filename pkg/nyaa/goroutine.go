package nyaa

import "context"

func GenTorrentsCh(ctx context.Context, torrents []Torrent) <-chan Torrent {
	torrCh := make(chan Torrent)

	go func() {
		defer close(torrCh)
		for _, tr := range torrents {
			select {
			case <-ctx.Done():
			case torrCh <- tr:
			}
		}
	}()

	return torrCh
}
