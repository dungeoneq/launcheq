package client

import (
	"bytes"
	"fmt"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"github.com/c2h5oh/datasize"
)

// Torrent downloads the torrent
func (c *Client) Torrent() error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = "."
	cfg.Debug = false
	cfg.Seed = false
	torrentClient, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("newClient: %w", err)
	}

	data, err := torrentContent.ReadFile("rof2.torrent")
	if err != nil {
		return fmt.Errorf("readFile %s: %w", "rof2.torrent", err)
	}

	mi, err := metainfo.Load(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("metainfo load: %w", err)
	}
	tr, err := torrentClient.AddTorrent(mi)
	if err != nil {
		return fmt.Errorf("addTorrent: %w", err)
	}

	start := time.Now()

	<-tr.GotInfo()

	defer tr.Drop()
	go func() {
		tick := time.NewTicker(6 * time.Second)

		for {
			select {
			case <-tick.C:
				st := tr.Stats()

				dataRate := (datasize.ByteSize(float64(st.BytesRead.Int64())/time.Since(start).Seconds()) * datasize.B)
				remainingTime := float64(tr.Info().TotalLength()) / float64(dataRate)

				totalPercent := float64(tr.BytesCompleted()) / float64(tr.Info().TotalLength()) * float64(100)

				fmt.Printf("peers: %d, seeders: %d, %s/s %0.2f%% of %s, ETA %0.1f minutes\n",
					st.ActivePeers,
					st.ConnectedSeeders,
					dataRate.HR(),
					totalPercent,
					(datasize.ByteSize(tr.Info().TotalLength()) * datasize.B).HR(),
					remainingTime/60)
			case <-tr.Closed():
				return
			}
		}
	}()
	c.logf("Downloading %s via Torrent", tr.Name())
	tr.DownloadAll()
	torrentClient.WaitAll()

	err = c.CopyBackup()
	if err != nil {
		return fmt.Errorf("copyBackup: %w", err)
	}

	fmt.Printf("Finished in %0.2f seconds\n", time.Since(start).Seconds())

	return nil
}
