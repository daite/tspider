package ktorrent

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/daite/tspider/common"
)

// TShare struct is for TShare torrent web site
type TShare struct {
	Name        string
	Keyword     string
	SearchURL   string
	ScrapedData *sync.Map
}

// Initialize method set keyword and URL based on default url
func (t *TShare) initialize(keyword string) {
	t.Keyword = keyword
	t.Name = "tshare"
	t.SearchURL = common.TorrentURL[t.Name] + "/bbs/search.php?sfl=wr_content&stx=" + url.QueryEscape(t.Keyword)
}

// Crawl torrent data from web site
func (t *TShare) Crawl(keyword string) map[string]string {
	t.initialize(keyword)
	data := t.getData(t.SearchURL)
	m := map[string]string{}
	data.Range(
		func(key, value interface{}) bool {
			m[fmt.Sprint(key)] = fmt.Sprint(value)
			return true
		})
	return m
}

// GetData method returns map(title, bbs url)
func (t *TShare) getData(url string) *sync.Map {
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
	doc.Find("li.list-item-row a").Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			title := strings.TrimSpace(s.Find("h1").Text())
			link, _ := s.Attr("href")
			magnet := t.GetMagnet(link)
			m.Store(title, magnet)
		}()
	})
	wg.Wait()
	t.ScrapedData = m
	return m
}

// GetMagnet method returns torrent magnet
func (t *TShare) GetMagnet(url string) string {
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return "failed to fetch magnet"
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err.Error()
	}
	magnet, ok := doc.Find("td a").Attr("href")
	if !ok {
		// maybe subtitles for movies
		return "no magnet"
	}
	return magnet
}
