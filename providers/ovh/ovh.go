package ovh

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
	ShortName   = "ovh"
	FullName    = "OVHcloud"
	HostType    = "hosting"
	SourceURL   = "https://www.ovhcloud.com/"
	DownloadURL = "https://vps.ovh.net/ips.txt"
)

type OVH struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() OVH {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return OVH{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

func (o *OVH) FetchData() (data []byte, headers http.Header, status int, err error) {
	if o.DownloadURL == "" {
		o.DownloadURL = DownloadURL
	}

	return web.Request(o.Client, o.DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (o *OVH) Fetch() (prefixes []netip.Prefix, err error) {
	data, _, _, err := o.FetchData()
	if err != nil {
		return
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (prefixes []netip.Prefix, err error) {
	r := bytes.NewReader(data)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		p, perr := netip.ParsePrefix(line)
		if perr != nil {
			continue
		}
		prefixes = append(prefixes, p)
	}
	return
}
