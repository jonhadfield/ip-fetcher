package abuseipdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/prefix-fetcher/internal/web"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"time"
)

const (
	APIURL     = "https://api.abuseipdb.com/api/v2/blacklist"
	TimeFormat = "2006-01-02T15:04:05-07:00"
)

type AbuseIPDB struct {
	Client            *retryablehttp.Client
	APIURL            string
	APIKey            string
	ConfidenceMinimum int
	Limit             int64
}

type RawBlacklistDoc struct {
	Meta struct {
		GeneratedAt string `json:"generatedAt"`
	} `json:"meta"`
	Data []struct {
		IPAddress            string `json:"ipAddress"`
		CountryCode          string `json:"countryCode"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		LastReportedAt       string `json:"LastReportedAt"`
	} `json:"data"`
}

func New() AbuseIPDB {
	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()
	c.HTTPClient = rc
	c.RetryMax = 1

	return AbuseIPDB{
		APIURL: APIURL,
		Client: c,
	}
}

func (a *AbuseIPDB) FetchData() (data []byte, headers http.Header, status int, err error) {
	// get download url if not specified
	if a.APIURL == "" {
		a.APIURL = APIURL
	}

	inHeaders := http.Header{}
	inHeaders.Add("Key", a.APIKey)
	inHeaders.Add("Accept", "application/json")

	var reqUrl *url.URL

	if reqUrl, err = url.Parse(a.APIURL); err != nil {
		return
	}

	if a.ConfidenceMinimum != 0 {
		reqUrl.Query().Add("confidenceMinimum", strconv.Itoa(a.ConfidenceMinimum))
	}

	if a.Limit != 0 {
		reqUrl.Query().Add("limit", strconv.FormatInt(a.Limit, 10))
	}

	blackList, headers, statusCode, err := web.Request(a.Client, reqUrl.String(), http.MethodGet, inHeaders, []string{a.APIKey}, 10*time.Second)
	if err != nil {
		return
	}

	if statusCode >= 400 && statusCode < 500 {
		err = parseAPIErrorResponse(blackList)
	}

	return blackList, headers, statusCode, err
}

type Doc struct {
	GeneratedAt time.Time
	Records     []Record
}

func (a *AbuseIPDB) Fetch() (doc Doc, err error) {
	data, _, _, err := a.FetchData()
	if err != nil {
		return
	}

	return Parse(data)
}

func Parse(in []byte) (doc Doc, err error) {
	var rawBlackListDoc RawBlacklistDoc

	err = json.Unmarshal(in, &rawBlackListDoc)
	if err != nil {
		err = fmt.Errorf("%w", err)

		return
	}

	doc.GeneratedAt, err = time.Parse(TimeFormat, rawBlackListDoc.Meta.GeneratedAt)
	if err != nil {
		return
	}

	for _, e := range rawBlackListDoc.Data {
		var lastReported time.Time

		if lastReported, err = time.Parse(TimeFormat, e.LastReportedAt); err != nil {
			return
		}

		doc.Records = append(doc.Records, Record{
			IPAddress:            netip.MustParseAddr(e.IPAddress),
			CountryCode:          e.CountryCode,
			AbuseConfidenceScore: e.AbuseConfidenceScore,
			LastReportedAt:       lastReported,
		})
	}

	return
}

type APIError struct {
	Detail string `json:"detail"`
	Status int    `json:"status"`
}

type APIErrorResponse struct {
	Errors []APIError `json:"errors"`
}

func parseAPIErrorResponse(in []byte) (err error) {
	var apiErrorResponse APIErrorResponse
	if err = json.Unmarshal(in, &apiErrorResponse); err != nil {
		return err
	}

	numErrors := len(apiErrorResponse.Errors)

	if numErrors == 0 {
		return errors.New("unexpected error: api error response did not contain any errors or parsing was unsuccessful")
	}

	e := apiErrorResponse.Errors[0]
	errString := fmt.Sprintf("AbuseIPDB returned '%s' with status code %d", e.Detail, e.Status)

	if numErrors > 1 {
		errString += " - and additional errors"
	}

	return errors.New(errString)
}

type Record struct {
	IPAddress            netip.Addr `json:"ipAddress"`
	CountryCode          string     `json:"countryCode"`
	AbuseConfidenceScore int        `json:"abuseConfidenceScore"`
	LastReportedAt       time.Time  `json:"lastReportedAt"`
}
