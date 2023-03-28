package abuseipdb

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jonhadfield/prefix-fetcher/internal/pflog"
	"github.com/jonhadfield/prefix-fetcher/internal/web"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"time"
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
		return false, fmt.Errorf("exceeded number of allowed blacklist downloads in last 24 hours")
	}

	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	return false, nil
}

func New() AbuseIPDB {
	pflog.SetLogLevel()

	rc := &http.Client{Transport: &http.Transport{}}
	c := retryablehttp.NewClient()

	if logrus.GetLevel() < logrus.DebugLevel {
		c.Logger = nil
	}

	c.CheckRetry = retryPolicy
	c.HTTPClient = rc
	c.RetryMax = 1

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

// constructReqUrl builds the AbuseIPDB API request URL with the provided values
func constructReqUrl(apiURL string, confidenceMinimum int, limit int64) (reqUrl *url.URL, err error) {
	if apiURL == "" {
		apiURL = APIURL
	}

	if reqUrl, err = url.Parse(apiURL); err != nil {
		return
	}

	q := reqUrl.Query()

	if confidenceMinimum != 0 {
		q.Add("confidenceMinimum", strconv.Itoa(confidenceMinimum))
	}

	if limit != 0 {
		q.Add("limit", strconv.FormatInt(limit, 10))
	}

	reqUrl.RawQuery = q.Encode()

	return reqUrl, nil
}

func (a *AbuseIPDB) FetchData() (data []byte, status int, err error) {
	reqUrl, err := constructReqUrl(a.APIURL, a.ConfidenceMinimum, a.Limit)
	if err != nil {
		return
	}

	reqHeaders := http.Header{}
	reqHeaders.Add("Key", a.APIKey)
	reqHeaders.Add("Accept", "application/json")

	blackList, _, statusCode, err := web.Request(a.Client, reqUrl.String(), http.MethodGet, reqHeaders, []string{a.APIKey}, 10*time.Second)
	if statusCode == 0 && err != nil {
		return
	}

	if len(blackList) == 0 {
		err = fmt.Errorf("empty response from %s api with http status code %d", ModuleName, statusCode)

		return
	}

	if statusCode >= 400 && statusCode < 500 {
		err = parseAPIErrorResponse(blackList)
	}

	return blackList, statusCode, err
}

type Doc struct {
	GeneratedAt time.Time
	Records     []Record
}

func (a *AbuseIPDB) Fetch() (doc Doc, err error) {
	data, status, err := a.FetchData()

	logrus.Debugf("abuseipdb | data len: %d FetchData status: %d", len(data), status)

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
