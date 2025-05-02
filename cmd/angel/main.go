package main

import (
	"fmt"
	"os"

	"github.com/daite/angel/common"
	"github.com/daite/angel/jtorrent"
	"github.com/daite/angel/ktorrent"
	"github.com/urfave/cli/v2"
)

var version = "0.9.0"

func main() {
	app := &cli.App{
		Name:    "angel",
		Usage:   "search torrent magnet!",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "choose torrent sites (kr or jp)",
			},
		},
		Action: func(c *cli.Context) error {
			// Handle the -V flag to print the version
			if c.Bool("print-version") {
				fmt.Printf("%s version %s\n", c.App.Name, c.App.Version)
				return nil
			}

			// Handle the case where no arguments are provided
			if c.Args().Len() == 0 {
				cli.ShowAppHelp(c)
				fmt.Printf("\nVersion: %s\n", c.App.Version)
				return nil
			}

			keyword := "동상이몽2"
			if c.NArg() > 0 {
				keyword = c.Args().Get(0)
			}

			if c.String("lang") == "kr" {

				sites := []common.Scraping{
					&ktorrent.TorrentTop{},
				}
				sites = common.GetAvailableSites(sites)
				fmt.Printf("[*] Angel found %d available site(s) ...\n", len(sites))
				data := common.CollectData(sites, keyword)
				common.PrintData(data)
			} else {
				sites := []common.ScrapingEx{
					&jtorrent.Nyaa{},
					&jtorrent.SuKeBe{},
				}
				sites = common.GetAvailableSitesEx(sites)
				fmt.Printf("[*] Angel found %d available site(s) ...\n", len(sites))
				data := common.CollectDataEx(sites, keyword)
				common.PrintDataEx(data)
			}
			return nil
		},
	}
	_ = app.Run(os.Args)
}
