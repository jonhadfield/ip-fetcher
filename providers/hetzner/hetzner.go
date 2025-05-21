package hetzner

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
	ShortName   = "hetzner"
	FullName    = "Hetzner"
	HostType    = "hosting"
	SourceURL   = "https://www.hetzner.com/"
	DownloadURL = "https://raw.githubusercontent.com/hetzneronline/ips/main/hetzner-all.txt"
)

type Hetzner struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() Hetzner {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Hetzner{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

func (h *Hetzner) FetchData() (data []byte, headers http.Header, status int, err error) {
	if h.DownloadURL == "" {
		h.DownloadURL = DownloadURL
	}

	return web.Request(h.Client, h.DownloadURL, http.MethodGet, nil, nil, 10*time.Second)
}

func (h *Hetzner) Fetch() (prefixes []netip.Prefix, err error) {
	data, _, _, err := h.FetchData()
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
		var prefix netip.Prefix
		prefix, err = netip.ParsePrefix(line)
		if err != nil {
			continue
		}
		prefixes = append(prefixes, prefix)
	}
	return
}
