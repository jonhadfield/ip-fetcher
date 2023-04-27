package cloudflare

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jonhadfield/prefix-fetcher/internal/pflog"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/prefix-fetcher/internal/web"
)

const (
	defaultIPv4URL           = "https://www.cloudflare.com/ips-v4"
	defaultIPv6URL           = "https://www.cloudflare.com/ips-v6"
	downloadedFileTimeFormat = "2006-01-02T15:04:05.999999"
)

func New() Cloudflare {
	pflog.SetLogLevel()

	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	c.HTTPClient = rc
	c.RetryMax = 1

	return Cloudflare{
		IPv4DownloadURL: defaultIPv4URL,
		IPv6DownloadURL: defaultIPv6URL,
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
		cf.IPv4DownloadURL = defaultIPv4URL
	}

	return web.Request(cf.Client, cf.IPv4DownloadURL, http.MethodGet, nil, nil, 10*time.Second)
}

func (cf *Cloudflare) FetchIPv6Data() (data []byte, headers http.Header, status int, err error) {
	if cf.IPv6DownloadURL == "" {
		cf.IPv6DownloadURL = defaultIPv6URL
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
		fmt.Println(scanner.Text()) // Println will add back the final '\n'

		var prefix netip.Prefix

		prefix, err = netip.ParsePrefix(scanner.Text())
		if err != nil {
			return
		}

		prefixes = append(prefixes, prefix)
	}

	return
}
