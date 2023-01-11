package gcp

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	u, err := url.Parse(DownloadURL)
	require.NoError(t, err)
	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		// SetHeader("Etag", "cd5e4f079775994d8e49f63ae9a84065").
		File("testdata/cloud.json")

	rc := retryablehttp.NewClient()
	ac := New()
	ac.Client = rc
	gock.InterceptClient(rc.HTTPClient)

	doc, err := ac.Fetch()
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
}
