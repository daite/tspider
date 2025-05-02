![example workflow name](https://github.com/daite/angel/workflows/Go/badge.svg)
 [![GoDoc](https://godoc.org/github.com/daite/angel?status.png)](http://godoc.org/github.com/daite/angel)
# 1. Torrent sites
* [추천 토렌트 사이트](http://jaewook.net/archives/2613)
# 2. Reference
* [cli](https://github.com/urfave/cli/blob/master/docs/v2/manual.md)
* [css selector](https://www.w3schools.com/cssref/css_selectors.asp)
* [tablewriter](https://github.com/olekukonko/tablewriter)
* [2d array sort](https://stackoverflow.com/questions/42629541/go-lang-sort-a-2d-array)
* [go channel](https://tour.golang.org/concurrency/4)
* [go module](https://blog.golang.org/using-go-modules)
* [goquery documentation](https://pkg.go.dev/github.com/PuerkitoBio/goquery)
* [css test env](https://try.jsoup.org/)
* [go binary size reduce](https://stackoverflow.com/questions/3861634/how-to-reduce-compiled-file-size)
* [go sync map to map](https://stackoverflow.com/questions/58995416/how-to-pretty-print-the-contents-of-a-sync-map)
* [go race detector](https://golang.org/doc/articles/race_detector.html)
* [json validator](https://jsonformatter.curiousconcept.com/)
* [json to map](https://gist.github.com/cuixin/f10cea0f8639454acdfbc0c9cdced764)
* [regex square blanket](https://stackoverflow.com/questions/928072/whats-the-regular-expression-that-matches-a-square-bracket)
* [online regex](https://regex101.com/)
* [codecoverage](https://codecov.io/gh/daite/)
* [cross compile](https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04)
# 3. ToDo
 - fetch torrent urls from [추천 토렌트 사이트](http://jaewook.net/archives/2613) but it is not a good approach.
# 4. Sample code
<details>
<summary>Click to toggle contents of `code`</summary>

```go
// Function to update the URL with a maximum number of retries
func updateTorrentURL(key string, url string, maxRetries int, wg *sync.WaitGroup, resultChan chan<- struct {
	key string
	url string
}) {
	defer wg.Done()
	for i := 0; i < maxRetries; i++ {
		if checkURL(url) {
			resultChan <- struct {
				key string
				url string
			}{key, url}
			return
		}
		url = incrementURL(url)
	}
	resultChan <- struct {
		key string
		url string
	}{key, ""} // Indicate failure with an empty string
}
```
</details>

# 5. How it works (feat.graphviz)
<img src="https://github.com/daite/angel/blob/main/resources/concept.png" width=70% height=70%>
