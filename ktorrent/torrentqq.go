package ktorrent

import (
	"fmt"
	"net/url"
	"regexp"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/daite/tspider/common"
)

// TorrentQQ struct is for TorrentQQ torrent web site
type TorrentQQ struct {
	Name        string
	Keyword     string
	SearchURL   string
	ScrapedData *sync.Map
}

// initialize method set keyword and URL based on default url
func (t *TorrentQQ) initialize(keyword string) {
	t.Keyword = keyword
	t.Name = "torrentqq"
	t.SearchURL = common.TorrentURL[t.Name] + "/search?q=" + url.QueryEscape(t.Keyword)
}

// Crawl torrent data from web site
func (t *TorrentQQ) Crawl(keyword string) map[string]string {
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
func (t *TorrentQQ) getData(url string) *sync.Map {
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
	doc.Find("a.subject").Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			title := s.Text()
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
func (t *TorrentQQ) GetMagnet(url string) string {
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return "failed to fetch magnet"
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err.Error()
	}
	re := regexp.MustCompile("[0-9,a-z]{40}")
	magnet := "no magnet"
	doc.Find("tr > td").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			text := re.FindString(s.Text())
			if len(text) == 40 {
				magnet = "magnet:?xt=urn:btih:" + text
			}
		}
	})
	return magnet
}
