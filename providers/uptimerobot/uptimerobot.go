package uptimerobot

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "uptimerobot"
	FullName  = "UptimeRobot"
	HostType  = "monitoring"
	SourceURL = "https://uptimerobot.com/help/locations/"
	// DownloadURL returns the monitoring probe addresses as a newline separated
	// list of bare IPv4 and IPv6 addresses.
	DownloadURL = "https://uptimerobot.com/inc/files/ips/IPv4andIPv6.txt"
)

type UptimeRobot struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() UptimeRobot {
	return UptimeRobot{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (u *UptimeRobot) FetchData() ([]byte, http.Header, int, error) {
	if u.DownloadURL == "" {
		u.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(u.Client, u.DownloadURL, http.MethodGet, nil, nil, u.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download uptimerobot addresses from %s. http status code: %d", u.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (u *UptimeRobot) Fetch() (Doc, error) {
	data, _, _, err := u.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	doc := Doc{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		entry := string(bytes.TrimSpace(scanner.Bytes()))
		if entry == "" {
			continue
		}

		prefix, ok := toPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse uptimerobot address: %s", entry)

			continue
		}

		if prefix.Addr().Is4() {
			doc.IPv4Prefixes = append(doc.IPv4Prefixes, prefix)

			continue
		}

		doc.IPv6Prefixes = append(doc.IPv6Prefixes, prefix)
	}

	if err := scanner.Err(); err != nil {
		return Doc{}, err
	}

	return doc, nil
}

// toPrefix accepts either a CIDR prefix or a bare IP address and returns a
// netip.Prefix (host prefixes are used for bare addresses).
func toPrefix(entry string) (netip.Prefix, bool) {
	if prefix, err := netip.ParsePrefix(entry); err == nil {
		return prefix, true
	}

	if addr, err := netip.ParseAddr(entry); err == nil {
		return netip.PrefixFrom(addr, addr.BitLen()), true
	}

	return netip.Prefix{}, false
}
