package digitalocean

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/prefix-fetcher/internal/web"
	"github.com/jonhadfield/prefix-fetcher/pflog"
	"github.com/jszwec/csvutil"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"net/netip"
	"time"
)

const (
	downloadURL         = "https://www.digitalocean.com/geo/google.csv"
	errFailedToDownload = "failed to download digital ocean prefixes document "
)

type DigitalOcean struct {
	Client      *retryablehttp.Client
	DownloadURL string
}

func New() DigitalOcean {
	pflog.SetLogLevel()
	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()
	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}
	c.HTTPClient = rc
	c.RetryMax = 1

	return DigitalOcean{
		DownloadURL: downloadURL,
		Client:      c,
	}
}

func (a *DigitalOcean) FetchData() (data []byte, headers http.Header, status int, err error) {
	// get download url if not specified
	if a.DownloadURL == "" {
		a.DownloadURL = downloadURL
	}

	data, headers, status, err = web.Request(a.Client, a.DownloadURL, http.MethodGet, nil, nil, 5*time.Second)
	if status >= 400 {
		return nil, nil, status, fmt.Errorf("failed to download prefixes. http status code: %d", status)
	}

	return data, headers, status, err
}

type Doc struct {
	LastModified time.Time
	ETag         string
	Records      []Record
}

func (a *DigitalOcean) Fetch() (doc Doc, err error) {
	data, headers, _, err := a.FetchData()
	if err != nil {
		return
	}

	records, err := Parse(data)
	if err != nil {
		return
	}

	doc.Records = records

	var etag string

	etags := headers.Values("etag")
	if len(etags) != 0 {
		etag = etags[0]
	}

	doc.ETag = etag

	var lastModifiedTime time.Time
	lastModifiedRaw := headers.Values("last-modified")
	if len(lastModifiedRaw) != 0 {
		if lastModifiedTime, err = time.Parse(time.RFC1123, lastModifiedRaw[0]); err != nil {
			return
		}
	}

	doc.LastModified = lastModifiedTime

	return doc, err
}

type Entry struct {
	Network     string `csv:"network,omitempty"`
	CountryCode string `csv:"countrycode,omitempty"`
	CityCode    string `csv:"citycode,omitempty"`
	CityName    string `csv:"cityname,omitempty"`
	ZipCode     string `csv:"zipcode,omitempty"`
}

func Parse(data []byte) (records []Record, err error) {
	reader := bytes.NewReader(data)
	csvReader := csv.NewReader(reader)
	doHeader, err := csvutil.Header(Entry{}, "csv")
	if err != nil {
		log.Fatal(err)
	}

	dec, err := csvutil.NewDecoder(csvReader, doHeader...)
	if err != nil {
		log.Fatal(err)
	}

Loop:
	for {
		var c Record
		err = dec.Decode(&c)
		switch err {
		case io.EOF:
			err = nil

			break Loop
		case nil:
			var pcn netip.Prefix

			pcn, err = netip.ParsePrefix(c.NetworkText)
			c.Network = pcn
			if err != nil {
				return records, err
			}

			records = append(records, c)
		default:
			return
		}
	}

	return
}

type Record struct {
	Network     netip.Prefix
	NetworkText string `csv:"network,omitempty"`
	CountryCode string `csv:"countrycode,omitempty"`
	CityCode    string `csv:"citycode,omitempty"`
	CityName    string `csv:"cityname,omitempty"`
	ZipCode     string `csv:"zipcode,omitempty"`
}

type CSVEntry struct {
	Network     string `csv:"network,omitempty"`
	CountryCode string `csv:"countrycode,omitempty"`
	CityCode    string `csv:"citycode,omitempty"`
	CityName    string `csv:"cityname,omitempty"`
	ZipCode     string `csv:"zipcode,omitempty"`
}
