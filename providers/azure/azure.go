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

func (a *Azure) FetchData() ([]byte, http.Header, int, error) {
	if a.DownloadURL == "" {
		// Microsoft rotates the dated snapshot URL on roughly a weekly cadence.
		// Scrape the download page for the current URL and fall back to the
		// last-known snapshot if discovery fails.
		discoveredURL, err := a.GetDownloadURL()
		if err != nil || discoveredURL == "" {
			a.DownloadURL = WorkaroundDownloadURL
		} else {
			a.DownloadURL = discoveredURL
		}
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
