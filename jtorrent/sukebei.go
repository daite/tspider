package jtorrent

import (
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/daite/tspider/common"
)

// SClient struct is for sending tasks
type SClient struct {
	title string
	link  string
}

// SData struct is for receiving final data
type SData struct {
	title string
	info  []string
}

var (
	sclients = make(chan SClient, 100)
	sdata    = make(chan SData, 100)
)

func screate(doc *goquery.Document, baseURL string) {
	doc.Find("a[href*=view]:last-child").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		link, _ := s.Attr("href")
		link = baseURL + link
		c := SClient{title, link}
		sclients <- c
	})
	close(sclients)
}

// SuKeBe struct is for sukebei torrent web site
type SuKeBe struct {
	Name        string
	Keyword     string
	SearchURL   string
	ScrapedData map[string][]string
}

// initialize method set keyword and URL based on default url
func (s *SuKeBe) initialize(keyword string) {
	s.Keyword = keyword
	s.Name = "sukebe"
	s.SearchURL = common.TorrentURL[s.Name] + "/?f=0&c=0_0&q=" + url.QueryEscape(s.Keyword)
}

// Crawl torrent data from web site
// NOTE: status code error: 429 429 Too Many Requests for goroutines
// Max concurrent request: 5
func (s *SuKeBe) Crawl(keyword string) map[string][]string {
	s.initialize(keyword)
	data := s.getData(s.SearchURL)
	if data == nil {
		return nil
	}
	return data
}

// GetData method returns map(title, bbs url)
func (s *SuKeBe) getData(url string) map[string][]string {
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return nil
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil
	}
	go screate(doc, common.TorrentURL[s.Name])
	s.makeWP(5)
	m := make(map[string][]string, 0)
	for d := range sdata {
		// Category 0
		// Time     1
		// Uploader 2
		// Seeder   3
		// Info     4
		// Leecher  5
		// FileSize 6
		// Snatch   7
		// Magnet   8
		uploader := d.info[2]
		seeder := d.info[3]
		leecher := d.info[5]
		snatch := d.info[7]
		fileSize := d.info[6]
		hash := d.info[8]
		folder := d.info[9]
		magnet := "magnet:?xt=urn:btih:" + hash
		info := []string{
			uploader, seeder, leecher,
			snatch, fileSize, magnet, folder,
		}
		title := common.RemoveNonAscII(d.title) + " _ " + hash[:5]
		m[title] = info
	}
	s.ScrapedData = m
	return m
}

// GetInfo method returns torrent info
func (s *SuKeBe) GetInfo(url string) []string {
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

func (s *SuKeBe) worker(wg *sync.WaitGroup) {
	for c := range sclients {
		title := c.title
		info := s.GetInfo(c.link)
		sdata <- SData{title, info}
	}
	wg.Done()
}

func (s *SuKeBe) makeWP(num int) {
	// Max concurrent should be "5"
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go s.worker(&wg)
	}
	wg.Wait()
	close(sdata)
}
