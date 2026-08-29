package statuscake

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "statuscake"
	FullName  = "StatusCake"
	HostType  = "monitoring"
	SourceURL = "https://www.statuscake.com/"
	// DownloadURL returns the test locations as a json object keyed by index,
	// each carrying the probe's address and, where it has one, its IPv6 address.
	DownloadURL = "https://app.statuscake.com/Workfloor/Locations.php?format=json"
)

type StatusCake struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() StatusCake {
	return StatusCake{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

// Location is a single test location. Not every location has an IPv6 address,
// so IPv6 is frequently empty.
type Location struct {
	GUID       string `json:"guid"`
	ServerCode string `json:"servercode"`
	Title      string `json:"title"`
	IP         string `json:"ip"`
	IPv6       string `json:"ipv6"`
	CountryISO string `json:"countryiso"`
	Status     string `json:"status"`
}

type Doc struct {
	Locations    []Location     `json:"locations" yaml:"locations"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

func (s *StatusCake) FetchData() ([]byte, http.Header, int, error) {
	if s.DownloadURL == "" {
		s.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(s.Client, s.DownloadURL, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download statuscake locations from %s. http status code: %d", s.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (s *StatusCake) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	// the document is keyed by an arbitrary index rather than being an array.
	var raw map[string]Location
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{Locations: make([]Location, 0, len(raw))}

	for _, location := range raw {
		doc.Locations = append(doc.Locations, location)

		addPrefix(&doc.IPv4Prefixes, location.IP)
		addPrefix(&doc.IPv6Prefixes, location.IPv6)
	}

	return doc, nil
}

// addPrefix appends the address as a host prefix, ignoring the empty string
// that locations without an IPv6 address carry.
func addPrefix(out *[]netip.Prefix, address string) {
	if address == "" {
		return
	}

	prefix, ok := iplist.ToPrefix(address)
	if !ok {
		logrus.Warnf("failed to parse statuscake address: %s", address)

		return
	}

	*out = append(*out, prefix)
}
