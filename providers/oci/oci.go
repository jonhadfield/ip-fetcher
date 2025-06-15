package oci

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/web"
)

const (
	ShortName                = "oci"
	FullName                 = "Oracle Cloud Infrastructure"
	HostType                 = "hosting"
	SourceURL                = "https://docs.oracle.com/en-us/iaas/Content/General/Concepts/addressranges.htm"
	DownloadURL              = "https://docs.oracle.com/iaas/tools/public_ip_ranges.json"
	downloadedFileTimeFormat = "2006-01-02T15:04:05.999999"
)

func New() OCI {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	return OCI{
		DownloadURL: DownloadURL,
		Client:      c,
	}
}

type OCI struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

type RawCIDR struct {
	CIDR string   `json:"cidr"`
	Tags []string `json:"tags"`
}

type CIDR struct {
	CIDR netip.Prefix `json:"cidr"`
	Tags []string     `json:"tags"`
}
type Region struct {
	Region string `json:"region"`
	CIDRS  []CIDR `json:"cidrs"`
}

type RawRegion struct {
	Region string    `json:"region"`
	CIDRS  []RawCIDR `json:"cidrs"`
	Tags   []string  `json:"tags"`
}

type RawDoc struct {
	LastUpdatedTimestamp string      `json:"last_updated_timestamp"`
	RawRegions           []RawRegion `json:"regions"`
}

func (doc *Doc) UnmarshalJSON(p []byte) error {
	var r RawDoc
	if err := json.Unmarshal(p, &r); err != nil {
		return err
	}

	ct, err := time.Parse(downloadedFileTimeFormat, r.LastUpdatedTimestamp)
	if err != nil {
		return err
	}

	var d Doc
	d.LastUpdatedTimestamp = ct.UTC()

	for _, rawRegion := range r.RawRegions {
		var region Region
		region.Region = rawRegion.Region

		var finalCIDRS []CIDR
		for _, rawCIDR := range rawRegion.CIDRS {
			var finalCIDR CIDR
			finalCIDR.Tags = rawCIDR.Tags

			var prefix netip.Prefix

			prefix, err = netip.ParsePrefix(rawCIDR.CIDR)
			if err != nil {
				return nil
			}

			finalCIDR.CIDR = prefix

			finalCIDRS = append(finalCIDRS, finalCIDR)
		}

		region.CIDRS = finalCIDRS

		d.Regions = append(d.Regions, region)
	}

	*doc = d
	return nil
}

func (ora *OCI) FetchData() ([]byte, http.Header, int, error) {
	if ora.DownloadURL == "" {
		ora.DownloadURL = DownloadURL
	}

	return web.Request(ora.Client, ora.DownloadURL, http.MethodGet, nil, nil, web.DefaultRequestTimeout)
}

func (ora *OCI) Fetch() (Doc, error) {
	data, _, _, err := ora.FetchData()
	if err != nil {
		return Doc{}, err
	}
	var doc Doc
	err = json.Unmarshal(data, &doc)

	return doc, err
}

type Doc struct {
	LastUpdatedTimestamp time.Time
	Regions              []Region
}
