package jtorrent

import (
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/daite/tspider/common"
)

// Client struct is for sending tasks
type Client struct {
	title string
	link  string
}

// Data struct is for receiving final data
type Data struct {
	title string
	info  []string
}

var (
	clients = make(chan Client, 100)
	data    = make(chan Data, 100)
)

func create(doc *goquery.Document, baseURL string) {
	doc.Find("a[href*=view]:last-child").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		link, _ := s.Attr("href")
		link = baseURL + link
		c := Client{title, link}
		clients <- c
	})
	close(clients)
}

// Nyaa struct is for Nyaa torrent web site
type Nyaa struct {
	Name        string
	Keyword     string
	SearchURL   string
	ScrapedData map[string][]string
}

// initialize method set keyword and URL based on default url
func (n *Nyaa) initialize(keyword string) {
	n.Keyword = keyword
	n.Name = "nyaa"
	n.SearchURL = common.TorrentURL[n.Name] + "/?f=0&c=0_0&q=" + url.QueryEscape(n.Keyword)
}

// Crawl torrent data from web site
// NOTE: status code error: 429 429 Too Many Requests for goroutines
// Max concurrent request: 5
func (n *Nyaa) Crawl(keyword string) map[string][]string {
	n.initialize(keyword)
	data := n.getData(n.SearchURL)
	if data == nil {
		return nil
	}
	return data
}

// GetData method returns map(title, bbs url)
func (n *Nyaa) getData(url string) map[string][]string {
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return nil
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil
	}
	go create(doc, common.TorrentURL[n.Name])
	n.makeWP(5)
	m := make(map[string][]string, 0)
	for d := range data {
		title := d.title
		// Category 0
		// Time     1
		// Uploader 2
		// Seeder   3
		// Info     4
		// Leecher  5
		// FileSize 6
		// Snatch   7
		// Magnet   8
		// Folder   9
		uploader := d.info[2]
		seeder := d.info[3]
		leecher := d.info[5]
		snatch := d.info[7]
		fileSize := d.info[6]
		hash := d.info[8]
		folder := d.info[9]
		magnet := "magnet:?xt=urn:btih:" + hash
		info := []string{
			uploader, seeder, leecher, snatch,
			fileSize, magnet, folder,
		}
		m[title] = info
	}
	n.ScrapedData = m
	return m
}

// GetInfo method returns torrent info
func (n *Nyaa) GetInfo(url string) []string {
	info := make([]string, 10)
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return info
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return info
	}
	info = doc.Find("div.col-md-5").Map(func(i int, s *goquery.Selection) string {
		return strings.TrimSpace(s.Text())
	})
	folder := "No"
	if _, ok := doc.Find("a.folder").Attr("class"); ok {
		folder = "Yes"
	}
	info = append(info, folder)
	return info
}

func (n *Nyaa) worker(wg *sync.WaitGroup) {
	for c := range clients {
		title := c.title
		info := n.GetInfo(c.link)
		data <- Data{title, info}
	}
	wg.Done()
}

func (n *Nyaa) makeWP(num int) {
	// Max concurrent should be "5"
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go n.worker(&wg)
	}
	wg.Wait()
	close(data)
}
