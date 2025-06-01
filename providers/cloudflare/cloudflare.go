package cloudflare

import (
	"bufio"
	"bytes"
	"net/http"
	"net/netip"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	ShortName      = "cloudflare"
	FullName       = "Cloudflare"
	HostType       = "cdn"
	SourceURL      = "https://www.cloudflare.com/en-gb/ips/"
	DefaultIPv4URL = "https://www.cloudflare.com/ips-v4"
	DefaultIPv6URL = "https://www.cloudflare.com/ips-v6"
)

func New() Cloudflare {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return Cloudflare{
		IPv4DownloadURL: DefaultIPv4URL,
		IPv6DownloadURL: DefaultIPv6URL,
		Client:          c,
	}
}

type Cloudflare struct {
	Client          *retryablehttp.Client
	IPv4DownloadURL string
	IPv6DownloadURL string
}

func (cf *Cloudflare) FetchIPv4Data() ([]byte, http.Header, int, error) {
	if cf.IPv4DownloadURL == "" {
		cf.IPv4DownloadURL = DefaultIPv4URL
	}

	return web.Request(cf.Client, cf.IPv4DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (cf *Cloudflare) FetchIPv6Data() ([]byte, http.Header, int, error) {
	if cf.IPv6DownloadURL == "" {
		cf.IPv6DownloadURL = DefaultIPv6URL
	}

	return web.Request(cf.Client, cf.IPv6DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (cf *Cloudflare) Fetch4() ([]netip.Prefix, error) {
	data, _, _, err := cf.FetchIPv4Data()
	if err != nil {
		return nil, err
	}

	return ProcessData(data)
}

func (cf *Cloudflare) Fetch6() ([]netip.Prefix, error) {
	data, _, _, err := cf.FetchIPv6Data()
	if err != nil {
		return nil, err
	}

	return ProcessData(data)
}

func (cf *Cloudflare) Fetch() ([]netip.Prefix, error) {
	p4, err := cf.Fetch4()
	if err != nil {
		return nil, err
	}

	p6, err := cf.Fetch6()
	if err != nil {
		return nil, err
	}

	return append(p4, p6...), nil
}

func ProcessData(data []byte) ([]netip.Prefix, error) {
	r := bytes.NewReader(data)

	scanner := bufio.NewScanner(r)
	var prefixes []netip.Prefix
	for scanner.Scan() {
		prefix, err := netip.ParsePrefix(scanner.Text())
		if err != nil {
			return nil, err
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}
