package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName             = "azure"
	FullName              = "Microsoft Azure"
	HostType              = "cloud"
	InitialURL            = "https://www.microsoft.com/en-gb/download/details.aspx?id=56519"
	WorkaroundDownloadURL = "https://download.microsoft.com/download/7/1/d/71d86715-5596-4529-9b13-da13a5de5b63/ServiceTags_Public_20260601.json"

	// DatedURLTemplate builds a snapshot URL from a publication date. Microsoft
	// names each snapshot ServiceTags_Public_YYYYMMDD.json under a stable path.
	DatedURLTemplate = "https://download.microsoft.com/download/7/1/d/71d86715-5596-4529-9b13-da13a5de5b63/ServiceTags_Public_%s.json"

	// dateLayout matches the date in DatedURLTemplate.
	dateLayout = "20060102"

	// weeksToProbe bounds how far back FindDownloadURL looks. Microsoft
	// publishes weekly and keeps roughly a fortnight of snapshots, so three
	// covers the retention window with margin while keeping the worst case,
	// where every probe misses, short enough to fall back quickly.
	weeksToProbe = 3

	// daysPerWeek is the snapshot cadence.
	daysPerWeek = 7

	errFailedToDownload = "failed to retrieve azure prefixes initial page"
)

type Azure struct {
	Client      *retryablehttp.Client
	InitialURL  string
	DownloadURL string
	Timeout     time.Duration

	// PageFetcher retrieves the download page during discovery. It defaults to
	// FetchPageWithCycleTLS and is exported so tests can replace it: cycletls
	// does not use the shared http.Client, so gock cannot intercept it and any
	// test exercising discovery would otherwise reach the live page.
	PageFetcher PageFetcher
}

// PageFetcher retrieves a page, returning its body and HTTP status.
type PageFetcher func(url string) (body string, status int, err error)

// FetchPageWithCycleTLS retrieves the download page using a spoofed TLS
// fingerprint, which the download site requires.
func FetchPageWithCycleTLS(url string) (string, int, error) {
	client := cycletls.Init()
	defer client.Close()

	response, err := client.Do(url, cycletls.Options{
		Body:      "",
		Ja3:       "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-51-57-47-53-10,0-23-65281-10-11-35-16-5-51-43-13-45-28-21,29-23-24-25-256-257,0",
		UserAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:87.0) Gecko/20100101 Firefox/87.0",
	}, "GET")
	if err != nil {
		return "", 0, err
	}

	return response.Body, response.Status, nil
}

func (a *Azure) ShortName() string {
	return ShortName
}

func (a *Azure) FullName() string {
	return FullName
}

func (a *Azure) HostType() string {
	return HostType
}

func (a *Azure) SourceURL() string {
	return InitialURL
}

func New() Azure {
	return Azure{
		InitialURL:  InitialURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
		PageFetcher: FetchPageWithCycleTLS,
	}
}

func (a *Azure) GetDownloadURL() (string, error) {
	if a.InitialURL == "" {
		a.InitialURL = InitialURL
	}

	if a.PageFetcher == nil {
		a.PageFetcher = FetchPageWithCycleTLS
	}

	body, status, err := a.PageFetcher(a.InitialURL)
	if err != nil {
		return "", errors.New(errFailedToDownload)
	}

	if status >= http.StatusBadRequest {
		return "", errors.New(errFailedToDownload)
	}

	return ParseDownloadURL(body), nil
}

// ParseDownloadURL extracts the first download.microsoft.com link from the
// download page, returning an empty string if the page holds none.
func ParseDownloadURL(body string) string {
	reATags := regexp.MustCompile("<a [^>]+>")
	reHRefs := regexp.MustCompile("href=\"[^\"]+\"")
	reDownloadURL := regexp.MustCompile("(http|https)://[^\"]+")

	for _, aTag := range reATags.FindAllString(body, -1) {
		for _, hrefMatch := range reHRefs.FindAllString(aTag, -1) {
			if !strings.Contains(hrefMatch, "download.microsoft.com/download/") {
				continue
			}

			if url := reDownloadURL.FindString(hrefMatch); url != "" {
				return url
			}
		}
	}

	return ""
}

// CandidateDates returns the snapshot publication dates to try, newest first.
// Microsoft publishes on Mondays, so the search starts at the most recent one.
//
// Order matters: the previous week's snapshot is still served for a while, so
// probing oldest first would happily return stale prefixes.
func CandidateDates(now time.Time) []time.Time {
	day := now.UTC()

	offset := (int(day.Weekday()) - int(time.Monday) + daysPerWeek) % daysPerWeek
	monday := day.AddDate(0, 0, -offset)

	dates := make([]time.Time, 0, weeksToProbe)
	for i := range weeksToProbe {
		dates = append(dates, monday.AddDate(0, 0, -daysPerWeek*i))
	}

	return dates
}

// FindDownloadURL locates the current snapshot by asking the file host for each
// candidate name directly, returning an empty string if none answers.
//
// This exists to avoid GetDownloadURL, which scrapes the download page and
// needs cycletls to do it: that costs around ten seconds, against roughly fifty
// milliseconds for a HEAD here. Only the download page demands a spoofed TLS
// fingerprint; the file host itself does not.
func (a *Azure) FindDownloadURL(now time.Time) string {
	probe := a.probeClient()

	for _, date := range CandidateDates(now) {
		url := fmt.Sprintf(DatedURLTemplate, date.Format(dateLayout))

		_, _, status, err := web.Request(probe, url, http.MethodHead, nil, nil, web.ShortRequestTimeout)
		if err == nil && status == http.StatusOK {
			return url
		}
	}

	return ""
}

// probeClient returns a client that does not retry, sharing the underlying
// transport so tests intercepting a.Client see the probes too.
//
// A missing snapshot is an answer, not a failure worth retrying: with the
// default policy a probe that cannot connect costs several seconds of backoff
// each, and three of those would delay the fallback longer than the scrape it
// is meant to avoid.
func (a *Azure) probeClient() *retryablehttp.Client {
	probe := retryablehttp.NewClient()
	probe.RetryMax = 0
	probe.Logger = nil

	if a.Client != nil {
		probe.HTTPClient = a.Client.HTTPClient
	}

	return probe
}

func (a *Azure) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		a.DownloadURL = a.resolveDownloadURL()
	}

	data, headers, status, err := web.Request(
		a.Client,
		a.DownloadURL,
		http.MethodGet,
		nil,
		nil,
		a.Timeout,
	)
	if status >= http.StatusBadRequest {
		return nil, nil, status, fmt.Errorf("failed to download prefixes. http status code: %d", status)
	}

	return data, headers, status, err
}

func (a *Azure) Fetch() (Doc, string, error) {
	data, headers, _, err := a.FetchData()
	if err != nil {
		return Doc{}, "", err
	}

	var doc Doc
	if err = json.Unmarshal(data, &doc); err != nil {
		return Doc{}, "", err
	}

	md5 := headers.Get(web.ContentMD5Header)

	return doc, md5, nil
}

type Doc struct {
	ChangeNumber int     `json:"changeNumber"`
	Cloud        string  `json:"cloud"`
	Values       []Value `json:"values"`
}

type Value struct {
	Name       string     `json:"name"`
	ID         string     `json:"id"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	ChangeNumber    int      `json:"changeNumber"`
	Region          string   `json:"region"`
	RegionID        int      `json:"regionId"`
	Platform        string   `json:"platform"`
	SystemService   string   `json:"systemService"`
	AddressPrefixes []string `json:"addressPrefixes"`
	NetworkFeatures []string `json:"networkFeatures"`
}

// resolveDownloadURL finds the current snapshot, cheapest option first.
//
// Microsoft rotates the dated snapshot weekly. Probing the names directly is
// fast; scraping the download page is the fallback for a naming or path change,
// and the pinned snapshot is the last resort.
func (a *Azure) resolveDownloadURL() string {
	if url := a.FindDownloadURL(time.Now()); url != "" {
		return url
	}

	discoveredURL, err := a.GetDownloadURL()
	if err != nil || discoveredURL == "" {
		return WorkaroundDownloadURL
	}

	return discoveredURL
}
