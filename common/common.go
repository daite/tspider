package common

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

// Scraping interface is for web scraping
type Scraping interface {
	Crawl(string) map[string]string
}

// ScrapingEx interface is for web scraping
type ScrapingEx interface {
	Crawl(string) map[string][]string
}

// variable for scraping
var (
	TorrentURL = map[string]string{
		"nyaa":       "https://nyaa.si",
		"sukebe":     "https://sukebei.nyaa.si",
		"torrenttop": "https://torrenttop152.com",
	}
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36"
)

// GetResponseFromURL returns *http.Response from url
func GetResponseFromURL(url string) (resp *http.Response, ok bool) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return resp, false
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err = client.Do(req)
	if err != nil {
		return resp, false
	}
	if resp.StatusCode != 200 {
		return resp, false
	}
	return resp, true
}

// CollectData function executes web scraping based on each scrapper
func CollectData(s []Scraping, keyword string) map[string]string {
	fmt.Println("[*] Angel is collecting data ...")
	var wg sync.WaitGroup
	ch := make(chan map[string]string, len(TorrentURL))
	for _, i := range s {
		wg.Add(1)
		go func(v Scraping) {
			defer wg.Done()
			r := v.Crawl(keyword)
			if r == nil {
				return
			}
			ch <- r
		}(i)
	}
	wg.Wait()
	close(ch)
	m := map[string]string{}
	for elem := range ch {
		for k, v := range elem {
			k = strings.Replace(k, " ", "_", -1)
			if v == "no magnet" {
				continue
			}
			m[k] = v
		}
	}
	return m
}

// CollectDataEx function executes web scraping based on each scrapper
func CollectDataEx(s []ScrapingEx, keyword string) map[string][]string {
	fmt.Println("[*] Angel is collecting data ...")
	var wg sync.WaitGroup
	ch := make(chan map[string][]string, len(TorrentURL))
	for _, i := range s {
		wg.Add(1)
		go func(v ScrapingEx) {
			defer wg.Done()
			r := v.Crawl(keyword)
			if r == nil {
				return
			}
			ch <- r
		}(i)
	}
	wg.Wait()
	close(ch)
	m := map[string][]string{}
	for elem := range ch {
		for k, v := range elem {
			k = strings.Replace(k, " ", "_", -1)
			m[k] = v
		}
	}
	return m
}

// PrintData function prints scraped data to console
func PrintData(data map[string]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Title", "Magnet"})
	matrix := [][]string{}
	for k, v := range data {
		matrix = append(matrix, []string{k, v})
	}
	sort.SliceStable(matrix, func(i, j int) bool { return matrix[i][0] > matrix[j][0] })
	for _, v := range matrix {
		table.Append(v)
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

// PrintDataEx function prints scraped data to console
func PrintDataEx(data map[string][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Title", "Uploader", "Seeder", "Leecher",
		"Snatch", "FileSize", "Magnet", "Folder",
	})
	for k, v := range data {
		m := make([]string, 0)
		m = append(m, k)
		for _, i := range v {
			m = append(m, i)
		}
		table.Append(m)
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

// URLJoin function join baseURL and relURL
func URLJoin(baseURL string, relURL string) string {
	u, err := url.Parse(relURL)
	if err != nil {
		log.Fatal(err)
	}
	base, err := url.Parse(baseURL + "/bbs/")
	if err != nil {
		log.Fatal(err)
	}
	return base.ResolveReference(u).String()
}

// CheckNetWorkFromURL function checks network status
func CheckNetWorkFromURL(url string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// GetAvailableSites function gets available torrent sites
func GetAvailableSites(oldItems []Scraping) []Scraping {
	fmt.Println("\r[*] Angel is checking available torrent sites ...")
	newItems := make([]Scraping, 0)
	items := []string{
		"torrenttop",
	}
	ch := make(chan int, len(items))
	var wg sync.WaitGroup
	for n, title := range items {
		wg.Add(1)
		go func(i int, t string) {
			defer wg.Done()
			ok := CheckNetWorkFromURL(TorrentURL[t])
			if ok {
				ch <- i
			}
		}(n, title)
	}
	wg.Wait()
	close(ch)
	for v := range ch {
		newItems = append(newItems, oldItems[v])
	}
	return newItems
}

// GetAvailableSitesEx function gets available torrent sites
func GetAvailableSitesEx(oldItems []ScrapingEx) []ScrapingEx {
	fmt.Println("[*] Angel is checking available torrent sites ...")
	newItems := make([]ScrapingEx, 0)
	items := []string{"nyaa", "sukebe"}
	ch := make(chan int, len(items))
	var wg sync.WaitGroup
	for n, title := range items {
		wg.Add(1)
		go func(i int, t string) {
			defer wg.Done()
			ok := CheckNetWorkFromURL(TorrentURL[t])
			if ok {
				ch <- i
			}
		}(n, title)
	}
	wg.Wait()
	close(ch)
	for v := range ch {
		newItems = append(newItems, oldItems[v])
	}
	return newItems
}

// RemoveNonAscII remove non-ASCII characters
func RemoveNonAscII(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}
