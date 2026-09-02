package tenable

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	ShortName = "tenable"
	FullName  = "Tenable Cloud Scanners"
	HostType  = "scanner"
	SourceURL = "https://docs.tenable.com/vulnerability-management/Content/Settings/Sensors/CloudSensors.htm"
	// DownloadURL returns the addresses Tenable's cloud scanners scan from, in
	// the same shape as the AWS ranges document, with the FedRAMP scanners kept
	// in members of their own.
	DownloadURL = "https://docs.tenable.com/ip-ranges/data.json"
)

type Tenable struct {
	Client      *retryablehttp.Client
	DownloadURL string
	Timeout     time.Duration
}

func New() Tenable {
	return Tenable{
		DownloadURL: DownloadURL,
		Client:      web.NewHTTPClientWithLogger(),
		Timeout:     web.DefaultRequestTimeout,
	}
}

type RawDoc struct {
	SyncToken           string          `json:"syncToken"`
	CreateDate          string          `json:"createDate"`
	Prefixes            []RawPrefix     `json:"prefixes"`
	IPv6Prefixes        []RawIPv6Prefix `json:"ipv6_prefixes"`
	FedRAMPPrefixes     []RawPrefix     `json:"fedramp_prefixes"`
	FedRAMPIPv6Prefixes []RawIPv6Prefix `json:"fedramp_ipv6_prefixes"`
}

type RawPrefix struct {
	IPPrefix           string `json:"ip_prefix"`
	Region             string `json:"region"`
	NetworkBorderGroup string `json:"network_border_group"`
	Service            string `json:"service"`
	SensorGroup        string `json:"sensor_group"`
}

type RawIPv6Prefix struct {
	IPv6Prefix         string `json:"ipv6_prefix"`
	Region             string `json:"region"`
	NetworkBorderGroup string `json:"network_border_group"`
	Service            string `json:"service"`
	SensorGroup        string `json:"sensor_group"`
}

// Prefix carries the sensor group alongside the region, as that is what names
// the scanners in the Tenable console.
type Prefix struct {
	IPPrefix           netip.Prefix `json:"ip_prefix"            yaml:"ip_prefix"`
	Region             string       `json:"region"               yaml:"region"`
	NetworkBorderGroup string       `json:"network_border_group" yaml:"network_border_group"`
	Service            string       `json:"service"              yaml:"service"`
	SensorGroup        string       `json:"sensor_group"         yaml:"sensor_group"`
}

// Doc keeps the FedRAMP scanners apart from the commercial ones, as only
// FedRAMP customers are scanned from them.
type Doc struct {
	SyncToken           string   `json:"syncToken"             yaml:"sync_token"`
	CreateDate          string   `json:"createDate"            yaml:"create_date"`
	Prefixes            []Prefix `json:"prefixes"              yaml:"prefixes"`
	IPv6Prefixes        []Prefix `json:"ipv6_prefixes"         yaml:"ipv6_prefixes"`
	FedRAMPPrefixes     []Prefix `json:"fedramp_prefixes"      yaml:"fedramp_prefixes"`
	FedRAMPIPv6Prefixes []Prefix `json:"fedramp_ipv6_prefixes" yaml:"fedramp_ipv6_prefixes"`
}

func (t *Tenable) FetchData() ([]byte, http.Header, int, error) {
	if t.DownloadURL == "" {
		t.DownloadURL = DownloadURL
	}

	data, headers, status, err := web.Request(t.Client, t.DownloadURL, http.MethodGet, nil, nil, t.Timeout)
	if err != nil {
		return nil, headers, status, err
	}

	if status >= http.StatusBadRequest {
		return nil, headers, status,
			fmt.Errorf("failed to download tenable ranges from %s. http status code: %d", t.DownloadURL, status)
	}

	return data, headers, status, nil
}

func (t *Tenable) Fetch() (Doc, error) {
	data, _, _, err := t.FetchData()
	if err != nil {
		return Doc{}, err
	}

	return ProcessData(data)
}

func ProcessData(data []byte) (Doc, error) {
	var raw RawDoc
	if err := json.Unmarshal(data, &raw); err != nil {
		return Doc{}, err
	}

	return Doc{
		SyncToken:           raw.SyncToken,
		CreateDate:          raw.CreateDate,
		Prefixes:            castPrefixes(raw.Prefixes),
		IPv6Prefixes:        castIPv6Prefixes(raw.IPv6Prefixes),
		FedRAMPPrefixes:     castPrefixes(raw.FedRAMPPrefixes),
		FedRAMPIPv6Prefixes: castIPv6Prefixes(raw.FedRAMPIPv6Prefixes),
	}, nil
}

func castPrefixes(in []RawPrefix) []Prefix {
	out := make([]Prefix, 0, len(in))

	for _, entry := range in {
		prefix, err := netip.ParsePrefix(entry.IPPrefix)
		if err != nil {
			logrus.Warnf("failed to parse tenable prefix: %s", entry.IPPrefix)

			continue
		}

		out = append(out, Prefix{
			IPPrefix:           prefix,
			Region:             entry.Region,
			NetworkBorderGroup: entry.NetworkBorderGroup,
			Service:            entry.Service,
			SensorGroup:        entry.SensorGroup,
		})
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func castIPv6Prefixes(in []RawIPv6Prefix) []Prefix {
	converted := make([]RawPrefix, 0, len(in))
	for _, entry := range in {
		converted = append(converted, RawPrefix{
			IPPrefix:           entry.IPv6Prefix,
			Region:             entry.Region,
			NetworkBorderGroup: entry.NetworkBorderGroup,
			Service:            entry.Service,
			SensorGroup:        entry.SensorGroup,
		})
	}

	return castPrefixes(converted)
}
