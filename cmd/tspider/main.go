package main

import (
	"fmt"
	"os"

	"github.com/daite/tspider/common"
	"github.com/daite/tspider/jtorrent"
	"github.com/daite/tspider/ktorrent"
	"github.com/urfave/cli/v2"
)

var version = "1.0.0"

func main() {
	app := &cli.App{
		Name:    "tspider",
		Usage:   "search torrent magnet links",
		Version: version,
		Commands: []*cli.Command{
			searchCommand(),
			doctorCommand(),
			configCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "choose torrent sites (kr or jp)",
			},
		},
		Action: func(c *cli.Context) error {
			// Default action: search if keyword provided
			if c.NArg() == 0 {
				return cli.ShowAppHelp(c)
			}
			return doSearch(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func searchCommand() *cli.Command {
	return &cli.Command{
		Name:      "search",
		Aliases:   []string{"s"},
		Usage:     "search for torrents",
		ArgsUsage: "<keyword>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "language filter: kr (Korean) or jp (Japanese)",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("please provide a search keyword")
			}
			return doSearch(c)
		},
	}
}

func doctorCommand() *cli.Command {
	return &cli.Command{
		Name:    "doctor",
		Aliases: []string{"d"},
		Usage:   "check availability of all torrent sites",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "check only sites for language: kr or jp",
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println("[*] Checking torrent site availability...")
			statuses := common.Doctor(c.String("lang"))
			common.PrintDoctorStatus(statuses)
			return nil
		},
	}
}

func configCommand() *cli.Command {
	return &cli.Command{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "manage site configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list all configured sites",
				Action: func(c *cli.Context) error {
					common.ListSites()
					return nil
				},
			},
			{
				Name:      "set-url",
				Usage:     "update a site's URL",
				ArgsUsage: "<site> <new-url>",
				Action: func(c *cli.Context) error {
					if c.NArg() < 2 {
						return fmt.Errorf("usage: angel config set-url <site> <new-url>")
					}
					site := c.Args().Get(0)
					url := c.Args().Get(1)
					if err := common.SetSiteURL(site, url); err != nil {
						return err
					}
					fmt.Printf("[+] Updated %s URL to: %s\n", site, url)
					return nil
				},
			},
			{
				Name:      "add",
				Usage:     "add a new site",
				ArgsUsage: "<name> <url> <language>",
				Action: func(c *cli.Context) error {
					if c.NArg() < 3 {
						return fmt.Errorf("usage: angel config add <name> <url> <language>\n  language: kr or jp")
					}
					name := c.Args().Get(0)
					url := c.Args().Get(1)
					lang := c.Args().Get(2)
					if lang != "kr" && lang != "jp" {
						return fmt.Errorf("language must be 'kr' or 'jp'")
					}
					if err := common.AddSite(name, url, lang); err != nil {
						return err
					}
					fmt.Printf("[+] Added site: %s (%s)\n", name, url)
					return nil
				},
			},
			{
				Name:      "remove",
				Usage:     "remove a site",
				ArgsUsage: "<site>",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please provide a site name")
					}
					name := c.Args().First()
					if err := common.RemoveSite(name); err != nil {
						return err
					}
					fmt.Printf("[+] Removed site: %s\n", name)
					return nil
				},
			},
			{
				Name:      "enable",
				Usage:     "enable a site",
				ArgsUsage: "<site>",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please provide a site name")
					}
					if err := common.EnableSite(c.Args().First(), true); err != nil {
						return err
					}
					fmt.Printf("[+] Enabled: %s\n", c.Args().First())
					return nil
				},
			},
			{
				Name:      "disable",
				Usage:     "disable a site",
				ArgsUsage: "<site>",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please provide a site name")
					}
					if err := common.EnableSite(c.Args().First(), false); err != nil {
						return err
					}
					fmt.Printf("[+] Disabled: %s\n", c.Args().First())
					return nil
				},
			},
			{
				Name:  "path",
				Usage: "show config file path",
				Action: func(c *cli.Context) error {
					fmt.Println(common.GetConfigPath())
					return nil
				},
			},
		},
	}
}

func doSearch(c *cli.Context) error {
	keyword := c.Args().First()
	if keyword == "" {
		return fmt.Errorf("please provide a search keyword")
	}

	lang := c.String("lang")

	if lang == "kr" {
		sites := []common.Scraping{
			&ktorrent.TorrentTop{},
		}
		sites, spinner := common.GetAvailableSites(sites)
		if len(sites) == 0 {
			spinner.Stop()
			fmt.Println("[!] No available sites. Use 'angel doctor' to check status.")
			return nil
		}
		data := common.CollectData(sites, keyword, spinner)
		spinner.StopWithMessage(fmt.Sprintf("Found %d result(s) from %d site(s)", len(data), len(sites)))
		common.PrintData(data)
	} else {
		sites := []common.ScrapingEx{
			&jtorrent.Nyaa{},
			&jtorrent.SuKeBe{},
		}
		sites, spinner := common.GetAvailableSitesEx(sites)
		if len(sites) == 0 {
			spinner.Stop()
			fmt.Println("[!] No available sites. Use 'angel doctor' to check status.")
			return nil
		}
		data := common.CollectDataEx(sites, keyword, spinner)
		spinner.StopWithMessage(fmt.Sprintf("Found %d result(s) from %d site(s)", len(data), len(sites)))
		common.PrintDataEx(data)
	}
	return nil
}
