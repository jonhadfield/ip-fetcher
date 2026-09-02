package updown

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"sort"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "updown"
	FullName  = "updown.io"
	HostType  = "monitoring"
	SourceURL = "https://updown.io/api"
	// DownloadURL returns the monitoring nodes as a json object keyed by node
	// name, each carrying the node's addresses and its location.
	DownloadURL = "https://updown.io/api/nodes"
)

type Updown struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Updown {
	return Updown{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// Node is a single monitoring node. Name is the key the document is indexed by,
// rather than a member of the node itself.
type Node struct {
	Name        string  `json:"name" yaml:"name"`
	IP          string  `json:"ip" yaml:"ip"`
	IPv6        string  `json:"ip6" yaml:"ip6"`
	City        string  `json:"city" yaml:"city"`
	Country     string  `json:"country" yaml:"country"`
	CountryCode string  `json:"country_code" yaml:"country_code"`
	Latitude    float64 `json:"lat" yaml:"lat"`
	Longitude   float64 `json:"lng" yaml:"lng"`
}

type Doc struct {
	Nodes        []Node         `json:"nodes" yaml:"nodes"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (u *Updown) FetchData() ([]byte, http.Header, int, error) {
	if u.DownloadURL == "" {
		u.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(u.Client, u.DownloadURL, http.MethodGet, nil, nil, u.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download updown nodes from %s. http status code: %d", u.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (u *Updown) Fetch() (Doc, error) {
	data, _, _, err := u.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw map[string]Node
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{Nodes: make([]Node, 0, len(raw))}

	for name, node := range raw {
		node.Name = name
		doc.Nodes = append(doc.Nodes, node)
	}

	// map iteration order is random, so sort to keep the document stable.
	sort.Slice(doc.Nodes, func(i, j int) bool { return doc.Nodes[i].Name < doc.Nodes[j].Name })

	for _, node := range doc.Nodes {
		addPrefix(&doc.IPv4Prefixes, node.IP)
		addPrefix(&doc.IPv6Prefixes, node.IPv6)
	}

	return doc, nil
}

// addPrefix appends the address as a host prefix, ignoring the empty string a
// node without an address of that family carries.
func addPrefix(out *[]netip.Prefix, address string) {
	if address == "" {
		return
	}

	prefix, ok := iplist.ToPrefix(address)
	if !ok {
		logrus.Warnf("failed to parse updown address: %s", address)

		return
	}

	*out = append(*out, prefix)
}
