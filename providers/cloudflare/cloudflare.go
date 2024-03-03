package cloudflare

import (
	"bufio"
	"bytes"
	"net/http"
	"net/netip"
	"time"

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

func (cf *Cloudflare) FetchIPv4Data() (data []byte, headers http.Header, status int, err error) {
	if cf.IPv4DownloadURL == "" {
		cf.IPv4DownloadURL = DefaultIPv4URL
	}

	return web.Request(cf.Client, cf.IPv4DownloadURL, http.MethodGet, nil, nil, 10*time.Second)
}

func (cf *Cloudflare) FetchIPv6Data() (data []byte, headers http.Header, status int, err error) {
	if cf.IPv6DownloadURL == "" {
		cf.IPv6DownloadURL = DefaultIPv6URL
	}

	return web.Request(cf.Client, cf.IPv6DownloadURL, http.MethodGet, nil, nil, 10*time.Second)
}

func (cf *Cloudflare) Fetch4() (prefixes []netip.Prefix, err error) {
	data, _, _, err := cf.FetchIPv4Data()
	if err != nil {
		return
	}

	return ProcessData(data)
}

func (cf *Cloudflare) Fetch6() (prefixes []netip.Prefix, err error) {
	data, _, _, err := cf.FetchIPv6Data()
	if err != nil {
		return
	}

	return ProcessData(data)
}

func (cf *Cloudflare) Fetch() (prefixes []netip.Prefix, err error) {
	p4, err := cf.Fetch4()
	if err != nil {
		return
	}

	p6, err := cf.Fetch6()
	if err != nil {
		return
	}

	return append(p4, p6...), err
}

func ProcessData(data []byte) (prefixes []netip.Prefix, err error) {
	r := bytes.NewReader(data)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// fmt.Println(scanner.Text()) // Println will add back the final '\n'

		var prefix netip.Prefix

		prefix, err = netip.ParsePrefix(scanner.Text())
		if err != nil {
			return
		}

		prefixes = append(prefixes, prefix)
	}

	return
}
