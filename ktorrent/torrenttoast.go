package ktorrent

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/daite/tspider/common"
)

// TorrentToast struct is for TorrentToast torrent web site
type TorrentToast struct {
	Name        string
	Keyword     string
	SearchURL   string
	ScrapedData *sync.Map
}

// initialize method set keyword and URL based on default url
func (t *TorrentToast) initialize(keyword string) {
	t.Keyword = keyword
	t.Name = "torrenttoast"
	t.SearchURL = common.TorrentURL[t.Name] + "/bbs/search.php?&stx=" + url.QueryEscape(t.Keyword)
}

// Crawl torrent data from web site
func (t *TorrentToast) Crawl(keyword string) map[string]string {
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
func (t *TorrentToast) getData(url string) *sync.Map {
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
	doc.Find("div.media-heading a").Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			title := strings.TrimSpace(s.Text())
			link, _ := s.Attr("href")
			link = strings.TrimSpace(common.URLJoin(common.TorrentURL[t.Name], link))
			magnet := t.GetMagnet(link)
			m.Store(title, magnet)
		}()
	})
	wg.Wait()
	t.ScrapedData = m
	return m
}

// GetMagnet method returns torrent magnet
func (t *TorrentToast) GetMagnet(url string) string {
	resp, ok := common.GetResponseFromURL(url)
	if !ok {
		return "failed to fetch magnet"
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err.Error()
	}
	magnet := "no magnet"
	doc.Find("a.list-group-item").Each(func(i int, s *goquery.Selection) {
		if m, _ := s.Attr("href"); strings.Contains(m, "magnet:?xt=urn:btih:") {
			magnet = m
		}
	})
	return magnet
}
