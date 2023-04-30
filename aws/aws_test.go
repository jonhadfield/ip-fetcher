package aws

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestGetIPListETag(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	testETag := "0338bd4dc4ba7a050b9124d333376fc7"

	gock.New(urlBase).
		Head(u.Path).
		MatchHeaders(map[string]string{"Accept": "application/json"}).
		Reply(200).
		SetHeader("Etag", "0338bd4dc4ba7a050b9124d333376fc7")

	ac := New()
	ac.DownloadURL = DownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	etag, err := ac.FetchETag()
	require.NoError(t, err)
	require.NotEmpty(t, etag)
	require.Equal(t, testETag, etag)
}

func TestDownloadIPList(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Accept": "application/json"}).
		Reply(200).
		SetHeader("Etag", "\"cd5e4f079775994d8e49f63ae9a84065\"").
		File("testdata/ip-ranges.json")

	ac := New()
	ac.DownloadURL = DownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	doc, etag, err := ac.Fetch()
	require.NoError(t, err)
	require.Equal(t, "cd5e4f079775994d8e49f63ae9a84065", etag)
	require.Len(t, doc.Prefixes, 28)
	require.Len(t, doc.IPv6Prefixes, 2)
}

func TestDownloadIPListWithoutQuotedEtag(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		MatchHeaders(map[string]string{"Accept": "application/json"}).
		Reply(200).
		SetHeader("Etag", "dd5e4f079775994d8e49f63ae9a84065").
		File("testdata/ip-ranges.json")

	ac := New()
	ac.DownloadURL = DownloadURL
	gock.InterceptClient(ac.Client.HTTPClient)

	_, etag, err := ac.Fetch()
	require.NoError(t, err)
	require.Equal(t, "dd5e4f079775994d8e49f63ae9a84065", etag)
}
