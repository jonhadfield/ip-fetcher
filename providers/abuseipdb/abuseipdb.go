package abuseipdb

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/ip-fetcher/internal/pflog"
	"github.com/jonhadfield/ip-fetcher/internal/web"
	"github.com/sirupsen/logrus"
)

const (
	APIURL     = "https://api.abuseipdb.com/api/v2/blacklist"
	ModuleName = "AbuseIPDB"
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

func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, v
			}

			if schemeErrorRe.MatchString(v.Error()) {
				return false, v
			}

			if notTrustedErrorRe.MatchString(v.Error()) {
				return false, v
			}
			if _, ok = v.Err.(x509.UnknownAuthorityError); ok {
				return false, v
			}
		}

		return true, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return false, errors.New("exceeded number of allowed blacklist downloads in last 24 hours")
	}

	if resp.StatusCode == 0 || (resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode != http.StatusNotImplemented) {
		return true, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	return false, nil
}

func New() AbuseIPDB {
	pflog.SetLogLevel()

	c := web.NewHTTPClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	c.CheckRetry = retryPolicy

	return AbuseIPDB{
		APIURL: APIURL,
		Client: c,
	}
}

var (
	redirectsErrorRe  = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe     = regexp.MustCompile(`unsupported protocol scheme`)
	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

func (a *AbuseIPDB) FetchData() ([]byte, http.Header, int, error) {
	// get download url if not specified
	if a.APIURL == "" {
		a.APIURL = APIURL
	}

	inHeaders := http.Header{}
	inHeaders.Add("Key", a.APIKey)
	inHeaders.Add("Accept", "application/json")

	reqUrl, err := url.Parse(a.APIURL)
	if err != nil {
		return nil, nil, 0, err
	}

	if a.ConfidenceMinimum != 0 {
		q := reqUrl.Query()
		q.Add("confidenceMinimum", strconv.Itoa(a.ConfidenceMinimum))
		reqUrl.RawQuery = q.Encode()
	}

	if a.Limit != 0 {
		q := reqUrl.Query()
		q.Add("limit", strconv.FormatInt(a.Limit, 10))
		reqUrl.RawQuery = q.Encode()
	}

	blackList, headers, statusCode, err := web.Request(a.Client, reqUrl.String(), http.MethodGet, inHeaders, []string{a.APIKey}, web.DefaultRequestTimeout)
	if statusCode == 0 && err != nil {
		return nil, nil, 0, err
	}

	if len(blackList) == 0 {
		return nil, nil, 0, fmt.Errorf("empty response from %s api with http status code %d", ModuleName, statusCode)
	}

	if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
		err = parseAPIErrorResponse(blackList)
	}
	return blackList, headers, statusCode, err
}

type Doc struct {
	GeneratedAt time.Time
	Records     []Record
}

func (a *AbuseIPDB) Fetch() (Doc, error) {
	data, _, status, err := a.FetchData()
	logrus.Debugf("abuseipdb | data len: %d FetchData status: %d", len(data), status)
	if err != nil {
		return Doc{}, err
	}

	return Parse(data)
}

func Parse(in []byte) (Doc, error) {
	var rawBlackListDoc RawBlacklistDoc
	err := json.Unmarshal(in, &rawBlackListDoc)
	if err != nil {
		return Doc{}, fmt.Errorf("%w", err)
	}

	doc := Doc{}
	doc.GeneratedAt, err = time.Parse(TimeFormat, rawBlackListDoc.Meta.GeneratedAt)
	if err != nil {
		return Doc{}, err
	}

	for _, e := range rawBlackListDoc.Data {
		var lastReported time.Time
		lastReported, err = time.Parse(TimeFormat, e.LastReportedAt)
		if err != nil {
			return Doc{}, err
		}

		doc.Records = append(doc.Records, Record{
			IPAddress:            netip.MustParseAddr(e.IPAddress),
			CountryCode:          e.CountryCode,
			AbuseConfidenceScore: e.AbuseConfidenceScore,
			LastReportedAt:       lastReported,
		})
	}

	return doc, nil
}

type APIError struct {
	Detail string `json:"detail"`
	Status int    `json:"status"`
}

type APIErrorResponse struct {
	Errors []APIError `json:"errors"`
}

func parseAPIErrorResponse(in []byte) error {
	var apiErrorResponse APIErrorResponse
	if err := json.Unmarshal(in, &apiErrorResponse); err != nil {
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
