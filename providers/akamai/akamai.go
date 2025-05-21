package akamai

import (
	"bufio"
	"bytes"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "akamai"
	FullName    = "Akamai"
	HostType    = "cdn"
	SourceURL   = "https://techdocs.akamai.com/"
	DownloadURL = "https://ip-ranges.akamai.com/"
)

type Akamai struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() Akamai {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Akamai{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

func (a *Akamai) FetchData() (data []byte, headers http.Header, status int, err error) {
	if a.DownloadURL == "" {
		a.DownloadURL = DownloadURL
	}

	return web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, 10*time.Second)
}

func (a *Akamai) Fetch() (prefixes []netip.Prefix, err error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (prefixes []netip.Prefix, err error) {
	r := bytes.NewReader(data)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var prefix netip.Prefix
		prefix, err = netip.ParsePrefix(scanner.Text())
		if err != nil {
			return
		}
		prefixes = append(prefixes, prefix)
	}
	return
}
