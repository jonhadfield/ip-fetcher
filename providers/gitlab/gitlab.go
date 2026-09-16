package gitlab

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName = "gitlab"
	FullName  = "GitLab"
	HostType  = "saas"
	SourceURL = "https://docs.gitlab.com/ee/user/gitlab_com/#ip-range"
	// DownloadURL is the GitLab.com docs page source. GitLab publishes its
	// Web/API webhook ranges only in prose, so the addresses are taken from
	// the page.
	DownloadURL = "https://gitlab.com/gitlab-org/gitlab/-/raw/master/doc/user/gitlab_com/_index.md"
)

// errNoAddresses is returned when the page carries no CIDRs, which means it
// has been restructured.
var errNoAddresses = errors.New("failed to find any addresses on the gitlab.com IP range page")

// addressRegexp matches CIDRs published in backticks on the IP range section.
var addressRegexp = regexp.MustCompile("`([0-9a-fA-F:.]+/[0-9]+)`")

type GitLab struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() GitLab {
	return GitLab{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type Doc struct {
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindAddresses returns the CIDRs the page publishes, in the order they
// appear, with duplicates dropped.
func FindAddresses(page []byte) ([]string, error) {
	var addresses []string

	seen := make(map[string]struct{})

	for _, match := range addressRegexp.FindAllSubmatch(page, -1) {
		entry := strings.TrimSpace(string(match[1]))

		if _, ok := iplist.ToPrefix(entry); !ok {
			continue
		}

		if _, ok := seen[entry]; ok {
			continue
		}

		seen[entry] = struct{}{}

		addresses = append(addresses, entry)
	}

	if len(addresses) == 0 {
		return nil, errNoAddresses
	}

	return addresses, nil
}

// FetchData returns the addresses as a newline separated list: the page itself
// carries documentation markup that is not worth publishing.
func (g *GitLab) FetchData() ([]byte, http.Header, int, error) {
	if g.DownloadURL == "" {
		g.DownloadURL = DownloadURL
	}

	page, headers, status, err := web.Request(g.Client, g.DownloadURL, http.MethodGet, nil, nil, g.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download gitlab addresses from %s. http status code: %d", g.DownloadURL, status)
	}

	addresses, err := FindAddresses(page)
	if err != nil {
		return nil, headers, status, err
	}

	return []byte(strings.Join(addresses, "\n") + "\n"), headers, status, nil
}

func (g *GitLab) Fetch() (Doc, error) {
	data, _, _, err := g.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	ipv4, ipv6, err := iplist.Parse(ShortName, data)
	if err != nil {
		return Doc{}, err
	}

	return Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}, nil
}
