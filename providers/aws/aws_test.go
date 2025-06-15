package aws_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/aws"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestGetIPListETag(t *testing.T) {
	u, err := url.Parse(aws.DownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	testETag := "0338bd4dc4ba7a050b9124d333376fc7"

	gock.New(urlBase).
		Head(u.Path).
		MatchHeaders(map[string]string{"Accept": "application/json"}).
		Reply(http.StatusOK).
		SetHeader("Etag", "0338bd4dc4ba7a050b9124d333376fc7")

	ac := aws.New()
	ac.DownloadURL = aws.DownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	etag, err := ac.FetchETag()
	require.NoError(t, err)
	require.NotEmpty(t, etag)
	require.Equal(t, testETag, etag)
}

func TestDownloadIPList(t *testing.T) {
	u, err := url.Parse(aws.DownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Accept": "application/json"}).
		Reply(http.StatusOK).
		SetHeader("Etag", "\"cd5e4f079775994d8e49f63ae9a84065\"").
		File("testdata/ip-ranges.json")

	ac := aws.New()
	ac.DownloadURL = aws.DownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, etag, err := ac.Fetch()
	require.NoError(t, err)
	require.Equal(t, "cd5e4f079775994d8e49f63ae9a84065", etag)
	require.Len(t, doc.Prefixes, 28)
	require.Len(t, doc.IPv6Prefixes, 2)
}

func TestDownloadIPListWithoutQuotedEtag(t *testing.T) {
	u, err := url.Parse(aws.DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Accept": "application/json"}).
		Reply(http.StatusOK).
		SetHeader("Etag", "dd5e4f079775994d8e49f63ae9a84065").
		File("testdata/ip-ranges.json")

	ac := aws.New()
	ac.DownloadURL = aws.DownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	_, etag, err := ac.Fetch()
	require.NoError(t, err)
	require.Equal(t, "dd5e4f079775994d8e49f63ae9a84065", etag)
}
