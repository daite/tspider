package common

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olekukonko/tablewriter"
)

// Spinner for progress animation
type Spinner struct {
	frames  []string
	current int
	message string
	start   time.Time
	total   int32
	done    int32
	stop    chan struct{}
	stopped chan struct{}
	mu      sync.Mutex
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		start:   time.Now(),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

// SetTotal sets the total number of tasks
func (s *Spinner) SetTotal(total int) {
	atomic.StoreInt32(&s.total, int32(total))
}

// IncrDone increments the done counter
func (s *Spinner) IncrDone() {
	atomic.AddInt32(&s.done, 1)
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		defer close(s.stopped)

		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				s.render()
				s.current = (s.current + 1) % len(s.frames)
			}
		}
	}()
}

func (s *Spinner) render() {
	s.mu.Lock()
	msg := s.message
	s.mu.Unlock()

	elapsed := time.Since(s.start)
	total := atomic.LoadInt32(&s.total)
	done := atomic.LoadInt32(&s.done)

	// Build status line
	frame := s.frames[s.current]
	elapsedStr := formatDuration(elapsed)

	var status string
	if total > 0 && done > 0 {
		// Calculate ETA
		avgTime := elapsed / time.Duration(done)
		remaining := time.Duration(total-done) * avgTime
		etaStr := formatDuration(remaining)
		status = fmt.Sprintf("\r%s %s [%d/%d] %s (ETA: %s)          ",
			frame, msg, done, total, elapsedStr, etaStr)
	} else if total > 0 {
		status = fmt.Sprintf("\r%s %s [0/%d] %s          ",
			frame, msg, total, elapsedStr)
	} else {
		status = fmt.Sprintf("\r%s %s %s          ",
			frame, msg, elapsedStr)
	}

	fmt.Print(status)
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	close(s.stop)
	<-s.stopped
	// Clear line
	fmt.Print("\r                                                              \r")
}

// StopWithMessage stops and prints final message
func (s *Spinner) StopWithMessage(msg string) {
	close(s.stop)
	<-s.stopped
	elapsed := formatDuration(time.Since(s.start))
	fmt.Printf("\r✓ %s (%s)                                    \n", msg, elapsed)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

// Scraping interface is for web scraping
type Scraping interface {
	Crawl(string) map[string]string
}

// ScrapingEx interface is for web scraping
type ScrapingEx interface {
	Crawl(string) map[string][]string
}

// SiteConfig holds configuration for a single torrent site
type SiteConfig struct {
	URL      string `json:"url"`
	Enabled  bool   `json:"enabled"`
	Language string `json:"language"` // "kr" or "jp"
}

// Config holds the application configuration
type Config struct {
	Sites     map[string]SiteConfig `json:"sites"`
	UserAgent string                `json:"user_agent"`
	Timeout   int                   `json:"timeout_seconds"`
}

var (
	// TorrentURL is the map of site names to URLs (loaded from config)
	TorrentURL = map[string]string{}
	// UserAgent for HTTP requests
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	// config holds the loaded configuration
	config     *Config
	configOnce sync.Once
	configPath string
)

// GetConfigPath returns the config file path
func GetConfigPath() string {
	if configPath != "" {
		return configPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "angel.json"
	}
	return filepath.Join(home, ".tspider.json")
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Sites: map[string]SiteConfig{
			// Korean sites
			"torrenttop":    {URL: "https://torrenttop152.com", Enabled: true, Language: "kr"},
			"torrentqq":     {URL: "https://torrentqq282.com", Enabled: false, Language: "kr"},
			"tshare":        {URL: "https://tshare.org", Enabled: false, Language: "kr"},
			"torrentmobile": {URL: "https://torrentmobile10.com", Enabled: false, Language: "kr"},
			"ktxtorrent":    {URL: "https://ktxtorrent.com", Enabled: false, Language: "kr"},
			"jujutorrent":   {URL: "https://jujutorrent.com", Enabled: false, Language: "kr"},
			"torrentgram":   {URL: "https://torrentgram.com", Enabled: false, Language: "kr"},
			"torrentmax":    {URL: "https://torrentmax.com", Enabled: false, Language: "kr"},
			"torrentrj":     {URL: "https://torrentrj.com", Enabled: false, Language: "kr"},
			"torrentsee":    {URL: "https://torrentsee.com", Enabled: false, Language: "kr"},
			"torrentsir":    {URL: "https://torrentsir.com", Enabled: false, Language: "kr"},
			"torrentsome":   {URL: "https://torrentsome.com", Enabled: false, Language: "kr"},
			"torrenttoast":  {URL: "https://torrenttoast.com", Enabled: false, Language: "kr"},
			"torrentwiz":    {URL: "https://torrentwiz.com", Enabled: false, Language: "kr"},
			"torrentj":      {URL: "https://torrentj.com", Enabled: false, Language: "kr"},
			"torrentview":   {URL: "https://torrentview.com", Enabled: false, Language: "kr"},
			"ttobogo":       {URL: "https://ttobogo.com", Enabled: false, Language: "kr"},
			// Japanese sites
			"nyaa":   {URL: "https://nyaa.si", Enabled: true, Language: "jp"},
			"sukebe": {URL: "https://sukebei.nyaa.si", Enabled: true, Language: "jp"},
		},
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout:   10,
	}
}

// LoadConfig loads the configuration from file or creates default
func LoadConfig() *Config {
	configOnce.Do(func() {
		path := GetConfigPath()
		data, err := os.ReadFile(path)
		if err != nil {
			// File doesn't exist, create default
			config = DefaultConfig()
			SaveConfig(config)
			return
		}
		config = &Config{}
		if err := json.Unmarshal(data, config); err != nil {
			config = DefaultConfig()
		}
		// Update TorrentURL map
		for name, site := range config.Sites {
			if site.Enabled {
				TorrentURL[name] = site.URL
			}
		}
		UserAgent = config.UserAgent
	})
	return config
}

// SaveConfig saves the configuration to file
func SaveConfig(c *Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	path := GetConfigPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	config = c
	// Update TorrentURL map
	TorrentURL = make(map[string]string)
	for name, site := range c.Sites {
		if site.Enabled {
			TorrentURL[name] = site.URL
		}
	}
	return nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	if config == nil {
		LoadConfig()
	}
	return config
}

// SetSiteURL updates a site's URL
func SetSiteURL(name, url string) error {
	c := GetConfig()
	site, exists := c.Sites[name]
	if !exists {
		return fmt.Errorf("site '%s' not found. Use 'angel config add' to add new sites", name)
	}
	site.URL = url
	c.Sites[name] = site
	return SaveConfig(c)
}

// AddSite adds a new site configuration
func AddSite(name, url, language string) error {
	c := GetConfig()
	if _, exists := c.Sites[name]; exists {
		return fmt.Errorf("site '%s' already exists. Use 'angel config set-url' to update URL", name)
	}
	c.Sites[name] = SiteConfig{
		URL:      url,
		Enabled:  true,
		Language: language,
	}
	return SaveConfig(c)
}

// EnableSite enables or disables a site
func EnableSite(name string, enabled bool) error {
	c := GetConfig()
	site, exists := c.Sites[name]
	if !exists {
		return fmt.Errorf("site '%s' not found", name)
	}
	site.Enabled = enabled
	c.Sites[name] = site
	return SaveConfig(c)
}

// RemoveSite removes a site from configuration
func RemoveSite(name string) error {
	c := GetConfig()
	if _, exists := c.Sites[name]; !exists {
		return fmt.Errorf("site '%s' not found", name)
	}
	delete(c.Sites, name)
	return SaveConfig(c)
}

// GetEnabledSites returns all enabled sites for a language
func GetEnabledSites(language string) map[string]SiteConfig {
	c := GetConfig()
	result := make(map[string]SiteConfig)
	for name, site := range c.Sites {
		if site.Enabled && site.Language == language {
			result[name] = site
		}
	}
	return result
}

// SiteStatus represents the health status of a site
type SiteStatus struct {
	Name      string
	URL       string
	Available bool
	Latency   time.Duration
	Error     string
	Language  string
	Enabled   bool
}

// Doctor checks all configured sites and returns their status
func Doctor(language string) []SiteStatus {
	c := GetConfig()
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []SiteStatus
	)

	for name, site := range c.Sites {
		if language != "" && site.Language != language {
			continue
		}
		wg.Add(1)
		go func(n string, s SiteConfig) {
			defer wg.Done()
			status := SiteStatus{
				Name:     n,
				URL:      s.URL,
				Language: s.Language,
				Enabled:  s.Enabled,
			}

			client := &http.Client{Timeout: time.Duration(c.Timeout) * time.Second}
			req, err := http.NewRequest("GET", s.URL, nil)
			if err != nil {
				status.Error = err.Error()
				mu.Lock()
				results = append(results, status)
				mu.Unlock()
				return
			}
			req.Header.Set("User-Agent", c.UserAgent)

			start := time.Now()
			resp, err := client.Do(req)
			status.Latency = time.Since(start)

			if err != nil {
				status.Error = err.Error()
			} else {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					status.Available = true
				} else {
					status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
				}
			}

			mu.Lock()
			results = append(results, status)
			mu.Unlock()
		}(name, site)
	}

	wg.Wait()
	return results
}

// PrintDoctorStatus prints the doctor status in a formatted way
func PrintDoctorStatus(statuses []SiteStatus) {
	fmt.Println()
	fmt.Printf("%-15s %-40s %-8s %-8s %-10s %s\n", "SITE", "URL", "STATUS", "ENABLED", "LATENCY", "ERROR")
	fmt.Println(strings.Repeat("─", 100))

	// Sort by name
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	available := 0
	for _, s := range statuses {
		status := "DOWN"
		if s.Available {
			status = "OK"
			available++
		}
		enabled := "No"
		if s.Enabled {
			enabled = "Yes"
		}
		latency := fmt.Sprintf("%dms", s.Latency.Milliseconds())
		errMsg := s.Error
		if len(errMsg) > 25 {
			errMsg = errMsg[:22] + "..."
		}
		urlStr := s.URL
		if len(urlStr) > 38 {
			urlStr = urlStr[:35] + "..."
		}
		fmt.Printf("%-15s %-40s %-8s %-8s %-10s %s\n", s.Name, urlStr, status, enabled, latency, errMsg)
	}

	fmt.Println(strings.Repeat("─", 100))
	fmt.Printf("Total: %d sites, %d available, %d down\n", len(statuses), available, len(statuses)-available)
}

// ListSites prints all configured sites
func ListSites() {
	c := GetConfig()
	fmt.Printf("Config file: %s\n\n", GetConfigPath())

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Site", "URL", "Language", "Enabled"})

	// Sort sites by name
	names := make([]string, 0, len(c.Sites))
	for name := range c.Sites {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		site := c.Sites[name]
		enabled := "No"
		if site.Enabled {
			enabled = "Yes"
		}
		table.Append([]string{name, site.URL, site.Language, enabled})
	}
	table.Render()
}

// GetResponseFromURL returns *http.Response from url
func GetResponseFromURL(url string) (resp *http.Response, ok bool) {
	c := GetConfig()
	client := &http.Client{Timeout: time.Duration(c.Timeout) * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return resp, false
	}
	req.Header.Set("User-Agent", c.UserAgent)
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
func CollectData(s []Scraping, keyword string, spinner *Spinner) map[string]string {
	spinner.UpdateMessage("Searching")
	spinner.SetTotal(len(s))
	atomic.StoreInt32(&spinner.done, 0)

	var wg sync.WaitGroup
	ch := make(chan map[string]string, len(s))
	for _, i := range s {
		wg.Add(1)
		go func(v Scraping) {
			defer wg.Done()
			r := v.Crawl(keyword)
			spinner.IncrDone()
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
func CollectDataEx(s []ScrapingEx, keyword string, spinner *Spinner) map[string][]string {
	spinner.UpdateMessage("Searching")
	spinner.SetTotal(len(s))
	atomic.StoreInt32(&spinner.done, 0)

	var wg sync.WaitGroup
	ch := make(chan map[string][]string, len(s))
	for _, i := range s {
		wg.Add(1)
		go func(v ScrapingEx) {
			defer wg.Done()
			r := v.Crawl(keyword)
			spinner.IncrDone()
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
	c := GetConfig()
	client := &http.Client{Timeout: time.Duration(c.Timeout) * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// GetAvailableSites function gets available torrent sites
func GetAvailableSites(oldItems []Scraping) ([]Scraping, *Spinner) {
	spinner := NewSpinner("Checking sites")
	spinner.SetTotal(len(oldItems))
	spinner.Start()

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
			spinner.IncrDone()
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
	return newItems, spinner
}

// GetAvailableSitesEx function gets available torrent sites
func GetAvailableSitesEx(oldItems []ScrapingEx) ([]ScrapingEx, *Spinner) {
	spinner := NewSpinner("Checking sites")
	spinner.SetTotal(len(oldItems))
	spinner.Start()

	newItems := make([]ScrapingEx, 0)
	items := []string{"nyaa", "sukebe"}
	ch := make(chan int, len(items))
	var wg sync.WaitGroup
	for n, title := range items {
		wg.Add(1)
		go func(i int, t string) {
			defer wg.Done()
			ok := CheckNetWorkFromURL(TorrentURL[t])
			spinner.IncrDone()
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
	return newItems, spinner
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

func init() {
	// Load config on package initialization
	LoadConfig()
}
