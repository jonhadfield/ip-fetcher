package site24x7

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "site24x7"
	FullName  = "Site24x7"
	HostType  = "monitoring"
	SourceURL = "https://www.site24x7.com/multi-location-web-site-monitoring.html"
	// LocationsURL is the page the export links live on. The json export's own
	// address carries a generated token, so it is discovered from the page
	// rather than hardcoded.
	LocationsURL = SourceURL
)

// errNoExportURL is returned when the locations page carries no json export
// link, which means the page has been restructured.
var errNoExportURL = errors.New("failed to find the site24x7 json export link on the locations page")

// exportURLRegexp matches the json flavour of the export link. The same view is
// also offered as csv, tsv and xml, so the /json/ segment picks one of four.
var exportURLRegexp = regexp.MustCompile(`https://creatorapp\.zohopublic\.[a-z.]+/[^"'\s]+/json/[^"'\s]+`)

type Site24x7 struct {
	Client       *retryablehttp.Client
	LocationsURL string
	Timeout      time.Duration
}

func New() Site24x7 {
	return Site24x7{
		LocationsURL: LocationsURL,
		Client:       web.NewHTTPClientWithLogger(),
		Timeout:      web.DefaultRequestTimeout,
	}
}

// rawDoc is the upstream export, a single view holding one record per location
// address.
type rawDoc struct {
	Locations []rawLocation `json:"IP_Address_View"`
}

type rawLocation struct {
	ID   string `json:"ID"`
	City string `json:"City"`
	// Place is the location's country.
	Place string `json:"Place"`
	IP    string `json:"external_ip"`
	IPv6  string `json:"IPv6_Address_External"`
}

// Location is a single monitoring location. Most have no IPv6 address, so IPv6
// is frequently empty, and a handful publish a prefix rather than a single
// address.
type Location struct {
	ID      string `json:"id" yaml:"id"`
	City    string `json:"city" yaml:"city"`
	Country string `json:"country" yaml:"country"`
	IP      string `json:"ip" yaml:"ip"`
	IPv6    string `json:"ipv6" yaml:"ipv6"`
}

type Doc struct {
	Locations    []Location     `json:"locations" yaml:"locations"`
	IPv4Prefixes []netip.Prefix `json:"ipv4_prefixes" yaml:"ipv4_prefixes"`
	IPv6Prefixes []netip.Prefix `json:"ipv6_prefixes" yaml:"ipv6_prefixes"`
}

// FindExportURL returns the json export's address from the locations page.
func FindExportURL(page []byte) (string, error) {
	match := exportURLRegexp.Find(page)
	if match == nil {
		return "", errNoExportURL
	}

	// the link is taken from an href, so any entities have to be decoded.
	return html.UnescapeString(string(match)), nil
}

func (s *Site24x7) FetchData() ([]byte, http.Header, int, error) {
	if s.LocationsURL == "" {
		s.LocationsURL = LocationsURL
	}

	page, headers, status, err := s.request(s.LocationsURL, "locations page")
	if err != nil {
		return nil, headers, status, err
	}

	exportURL, err := FindExportURL(page)
	if err != nil {
		return nil, headers, status, err
	}

	logrus.Debugf("found site24x7 export url: %s", exportURL)

	return s.request(exportURL, "addresses")
}

func (s *Site24x7) request(url, what string) ([]byte, http.Header, int, error) {
	data, headers, status, err := web.Request(s.Client, url, http.MethodGet, nil, nil, s.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download site24x7 %s from %s. http status code: %d", what, url, status)
	}

	return data, headers, status, nil
}

func (s *Site24x7) Fetch() (Doc, error) {
	data, _, _, err := s.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw rawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	doc := Doc{Locations: make([]Location, 0, len(raw.Locations))}

	for _, location := range raw.Locations {
		// several records carry trailing whitespace around their addresses.
		entry := Location{
			ID:      strings.TrimSpace(location.ID),
			City:    strings.TrimSpace(location.City),
			Country: strings.TrimSpace(location.Place),
			IP:      strings.TrimSpace(location.IP),
			IPv6:    strings.TrimSpace(location.IPv6),
		}

		doc.Locations = append(doc.Locations, entry)

		addPrefix(&doc.IPv4Prefixes, entry.IP)
		addPrefix(&doc.IPv6Prefixes, entry.IPv6)
	}

	return doc, nil
}

// addPrefix appends the entry, which is either an address or a prefix, ignoring
// the empty string a location without an address of that family carries.
func addPrefix(out *[]netip.Prefix, entry string) {
	if entry == "" {
		return
	}

	prefix, ok := iplist.ToPrefix(entry)
	if !ok {
		logrus.Warnf("failed to parse site24x7 address: %s", entry)

		return
	}

	*out = append(*out, prefix)
}
