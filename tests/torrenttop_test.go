package tests

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGetDataFuncForTorrentTop(t *testing.T) {
	f, err := os.Open("../resources/torrenttop_search.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		log.Fatal(err)
	}
	got := make(map[string]string)
	doc.Find(".py-4.flex.flex-row.border-b.topic-item a").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Attr("title")
		title = strings.TrimSpace(title)
		link, _ := s.Attr("href")
		got[title] = link
	})
	want := map[string]string{
		"동상이몽2 너는 내운명.E177.201228.720p-NEXT": "/torrent/jro35vg.html",
	}
	if got["동상이몽2 너는 내운명.E177.201228.720p-NEXT"] != "/torrent/jro35vg.html" {
		t.Errorf("GetData() for TorrentTop = %q, want %q", got, want)
	}
}

func TestGetMagnetFuncForTorrentTop(t *testing.T) {
	f, err := os.Open("../resources/torrenttop_bbs.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		log.Fatal(err)
	}
	got, _ := doc.Find(".fas.fa-magnet + a").Attr("href")
	want := "magnet:?xt=urn:btih:6bb34701c93505114029e5c91a0e88a30c11703b"
	if got != want {
		t.Errorf("GetMagnet() for TorrentTop = %q, want %q", got, want)
	}
}
