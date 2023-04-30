package abuseipdb

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"net/netip"
	"net/url"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	data := []byte(`{"meta":{"generatedAt":"2022-07-06T21:18:45+00:00"},"data":[{"ipAddress":"104.255.199.22","countryCode":"US","abuseConfidenceScore":100,"lastReportedAt":"2022-07-06T21:17:02+00:00"},{"ipAddress":"59.49.78.12","countryCode":"CN","abuseConfidenceScore":100,"lastReportedAt":"2022-07-06T21:17:02+00:00"}]}`)
	doc, err := Parse(data)
	require.NoError(t, err)
	require.Len(t, doc.Records, 2)

	data = []byte(`invalid`)
	doc, err = Parse(data)
	require.Error(t, err)
	require.Empty(t, doc)
}

func TestFetchBlackListFailure(t *testing.T) {
	ac := New()
	ac.APIKey = "test-key"
	ac.APIURL = "https://example.com"
	u, err := url.Parse(ac.APIURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Key": "test-key", "Accept": "application/json"}).
		Reply(400)

	gock.InterceptClient(ac.Client.HTTPClient)

	_, err = ac.Fetch()
	require.Error(t, err)
	require.ErrorContains(t, err, "empty response")
}

func TestFetchBlackListData(t *testing.T) {
	ac := New()
	ac.APIKey = "test-key"
	ac.APIURL = "https://example.com"
	u, err := url.Parse(ac.APIURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Key": "test-key", "Accept": "application/json"}).
		Reply(200).
		File("testdata/blacklist")

	gock.InterceptClient(ac.Client.HTTPClient)

	data, _, status, err := ac.FetchData()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Equal(t, http.StatusOK, status)
}

func TestFetchBlackList(t *testing.T) {
	ac := New()
	ac.APIKey = "test-key"
	ac.APIURL = "https://example.com"
	u, err := url.Parse(ac.APIURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Key": "test-key", "Accept": "application/json"}).
		Reply(200).
		File("testdata/blacklist")

	gock.InterceptClient(ac.Client.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	expectedGeneratedAt, err := time.Parse(TimeFormat, "2022-07-06T21:18:45+00:00")
	require.NoError(t, err)
	require.Equal(t, expectedGeneratedAt, doc.GeneratedAt)
	require.NotEmpty(t, doc.Records)
	var found bool
	expectedAddr := netip.MustParseAddr("92.118.161.25")
	expectedReportedAt := "2022-07-06T21:17:00+00:00"
	for _, r := range doc.Records {
		if r.IPAddress == expectedAddr {
			if r.CountryCode == "US" {
				if r.AbuseConfidenceScore == 100 {
					expectedReportedAt, _ := time.Parse(TimeFormat, expectedReportedAt)
					if r.LastReportedAt.String() == expectedReportedAt.String() {
						found = true

						break
					}
				}
			}
		}
	}
	require.True(t, found)
}

func TestFetchBlackListDataIncorrectKey(t *testing.T) {
	ac := New()
	ac.APIKey = "incorrect-key"
	ac.APIURL = "https://example.com"
	u, err := url.Parse(ac.APIURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Key": "incorrect-key", "Accept": "application/json"}).
		Reply(401).
		JSON([]byte(`{"errors":[{"detail":"Authentication failed. Your API key is either missing, incorrect, or revoked. Note: The APIv2 key differs from the APIv1 key.","status":401}]}`))

	gock.InterceptClient(ac.Client.HTTPClient)

	data, _, status, err := ac.FetchData()
	require.Error(t, err)
	require.Contains(t, string(data), "Authentication failed.")
	require.Equal(t, http.StatusUnauthorized, status)
}
