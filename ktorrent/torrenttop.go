package ktorrent

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/daite/tspider/common"
)

// TorrentTop struct is for TorrentSee torrent web site
type TorrentTop struct {
	Name        string
	Keyword     string
	SearchURL   string
	ScrapedData *sync.Map
}

// initialize method set keyword and URL based on default url
func (t *TorrentTop) initialize(keyword string) {
	t.Keyword = keyword
	t.Name = "torrenttop"
	t.SearchURL = common.TorrentURL[t.Name] + "/search/index?keywords=" + url.QueryEscape(t.Keyword)
}

// Crawl torrent data from web site
func (t *TorrentTop) Crawl(keyword string) map[string]string {
	t.initialize(keyword)
	data := t.getData(t.SearchURL)
	if data == nil {
		return nil
	}
	m := map[string]string{}
	data.Range(
		func(key, value interface{}) bool {
			m[fmt.Sprint(key)] = fmt.Sprint(value)
			return true
		})
	return m
}

// GetData method returns map(title, bbs url)
func (t *TorrentTop) getData(url string) *sync.Map {
	var wg sync.WaitGroup
	m := &sync.Map{}

	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil
	}

	doc.Find(".topic-item a").Each(func(i int, s *goquery.Selection) {
		title, exists := s.Attr("title")
		href, linkOk := s.Attr("href")
		if !exists || !linkOk {
			return
		}

		wg.Add(1)
		go func(title, href string) {
			defer wg.Done()
			fullURL := strings.TrimSpace(common.URLJoin(common.TorrentURL[t.Name], href))
			magnet := t.GetMagnet(fullURL)
			m.Store(strings.TrimSpace(title), magnet)
		}(title, href)
	})

	wg.Wait()
	t.ScrapedData = m
	return m
}

// GetMagnet method returns torrent magnet
func (t *TorrentTop) GetMagnet(url string) string {
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return "failed to fetch magnet"
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Sprintf("parse error: %v", err)
	}

	magnet := ""
	doc.Find("i.fas.fa-magnet").Each(func(i int, s *goquery.Selection) {
		parent := s.Parent()
		parent.Find("a").EachWithBreak(func(i int, a *goquery.Selection) bool {
			href, exists := a.Attr("href")
			if exists && strings.HasPrefix(href, "magnet:?") {
				magnet = href
				return false
			}
			return true
		})
	})

	if magnet == "" {
		return "no magnet"
	}
	return magnet
}
