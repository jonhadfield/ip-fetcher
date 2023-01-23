package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jonhadfield/prefix-fetcher/pflog"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/prefix-fetcher/internal/web"
)

const (
	initialURL          = "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"
	errFailedToDownload = "failed to retrieve azure prefixes initial page"
)

type Azure struct {
	Client      *retryablehttp.Client
	InitialURL  string
	DownloadURL string
}

func New() Azure {
	pflog.SetLogLevel()
	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}
	c.HTTPClient = rc
	c.RetryMax = 1

	return Azure{
		InitialURL: initialURL,
		Client:     c,
	}
}

func (a *Azure) GetDownloadURL() (url string, err error) {
	if a.InitialURL == "" {
		a.InitialURL = initialURL
	}

	body, _, status, err := web.Request(a.Client, a.InitialURL, http.MethodGet, nil, nil, 10*time.Second)
	if err != nil {
		return
	}

	if status >= 400 {
		return url, errors.New(errFailedToDownload)
	}

	reATags := regexp.MustCompile("<a [^>]+>")

	aTags := reATags.FindAllString(string(body), -1)

	reHRefs := regexp.MustCompile("href=\"[^\"]+\"")

	var hrefs []string

	for _, href := range aTags {
		hrefMatches := reHRefs.FindAllString(href, -1)
		for _, hrefMatch := range hrefMatches {
			if strings.Contains(hrefMatch, "download.microsoft.com/download/") {
				hrefs = append(hrefs, hrefMatch)
			}
		}
	}

	reDownloadURL := regexp.MustCompile("(http|https)://[^\"]+")

	for _, href := range hrefs {
		url = reDownloadURL.FindString(href)
		if url != "" {
			break
		}
	}

	return
}

func (a *Azure) FetchData() (data []byte, headers http.Header, status int, err error) {
	// get download url if not specified
	if a.DownloadURL == "" {
		a.DownloadURL, err = a.GetDownloadURL()
		if err != nil {
			return
		}
	}

	data, headers, status, err = web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, 5*time.Second)
	if status >= 400 {
		return nil, nil, status, fmt.Errorf("failed to download prefixes. http status code: %d", status)
	}

	return data, headers, status, err
}

func (a *Azure) Fetch() (doc Doc, md5 string, err error) {
	data, headers, _, err := a.FetchData()
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &doc)
	if err != nil {
		return
	}

	md5 = headers.Get("Content-MD5")

	return
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
