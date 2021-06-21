package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/JohnRabin2357/nyaa-dl/pkg/client"
	"github.com/JohnRabin2357/nyaa-dl/pkg/config"
	"github.com/JohnRabin2357/nyaa-dl/pkg/download"
	"github.com/JohnRabin2357/nyaa-dl/pkg/nyaa"
	"github.com/urfave/cli/v2"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("[nyaa-dl] ")
}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("nyaa-dl failed: %s", err)
	}

	os.Exit(0)
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := &cli.App{
		Name:  "nyaa-dl",
		Usage: "Downloads .torrent files from nyaa.",
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "domain",
			Usage:    "The domain of a nyaa.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "config",
			Usage:    "The path of a configuration.",
			Required: true,
		},
		&cli.StringFlag{
			Name:        "output",
			Usage:       "The directory to save torrent files.",
			DefaultText: "./torrents",
		},
	}

	app.Action = func(c *cli.Context) error {
		log.Print("Launching Nyaa Downloader!")

		conf, err := config.ReadConfig(c.String("config"))
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		client := client.NewClient(&client.NewClientInput{
			Config:     conf,
			Nyaa:       nyaa.NewNyaa(c.String("domain")),
			Downloader: download.NewDownloader(c.String("output")),
		})

		start := time.Now()
		err = client.Run(c.Context)
		if err != nil {
			return fmt.Errorf("client error: %w", err)
		}

		log.Printf("Completed (after: %v).", time.Since(start))
		return nil
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		return fmt.Errorf("app error: %w", err)
	}
	return nil
}
