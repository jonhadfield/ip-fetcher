package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName             = "azure"
	FullName              = "Microsoft Azure"
	HostType              = "cloud"
	InitialURL            = "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"
	WorkaroundDownloadURL = "https://raw.githubusercontent.com/tobilg/public-cloud-provider-ip-ranges/main/data/providers/azure.json"

	errFailedToDownload = "failed to retrieve azure prefixes initial page"
)

type Azure struct {
	Client      *retryablehttp.Client
	InitialURL  string
	DownloadURL string
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
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Azure{
		InitialURL: InitialURL,
		Client:     c,
	}
}

func (a *Azure) GetDownloadURL() (url string, err error) {
	if a.InitialURL == "" {
		a.InitialURL = InitialURL
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
	// if a.DownloadURL == "" {
	// 	// hack whilst Akamai bot protection is in place
	// 	// a.DownloadURL = WorkaroundDownloadURL
	// 	a.DownloadURL, err = a.GetDownloadURL()
	// 	if err != nil {
	// 		return
	// 	}
	// }

	// hack whilst Akamai bot protection workaround not in place
	// provided by https://github.com/tobilg/public-cloud-provider-ip-ranges
	a.DownloadURL = WorkaroundDownloadURL

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
