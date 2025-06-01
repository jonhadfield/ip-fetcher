package zscaler

import (
	"bufio"
	"bytes"
	"net/http"
	"net/netip"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName   = "zscaler"
	FullName    = "Zscaler"
	HostType    = "security"
	SourceURL   = "https://www.zscaler.com"
	DownloadURL = "https://api.config.zscaler.com/zscaler.net/cenr/json"
)

type Zscaler struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() Zscaler {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Zscaler{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

func (z *Zscaler) FetchData() (data []byte, headers http.Header, status int, err error) {
	if z.DownloadURL == "" {
		z.DownloadURL = DownloadURL
	}

	return web.Request(z.Client, z.DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (z *Zscaler) Fetch() (prefixes []netip.Prefix, err error) {
	data, _, _, err := z.FetchData()
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
			continue
		}
		prefixes = append(prefixes, prefix)
	}
	return
}
