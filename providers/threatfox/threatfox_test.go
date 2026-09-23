package threatfox_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/threatfox"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(threatfox.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/threatfox.json")

	tf := threatfox.New()
	gock.InterceptClient(tf.Client.HTTPClient)

	doc, err := tf.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("155.103.69.239/32"))
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("8.137.111.232/32"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:db8::1/128"))
	require.Len(t, doc.Entries, 3)

	var found bool
	for _, e := range doc.Entries {
		if e.MalwarePrintable == "Remcos" {
			found = true
			require.Equal(t, "4550", e.Port)

			break
		}
	}
	require.True(t, found)
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/threatfox.json")
	require.NoError(t, err)

	doc, err := threatfox.ProcessData(data)
	require.NoError(t, err)
	require.Len(t, doc.IPv4Prefixes, 2)
	require.Len(t, doc.IPv6Prefixes, 1)
}

func TestProcessDataInvalidJSON(t *testing.T) {
	_, err := threatfox.ProcessData([]byte("not json"))
	require.Error(t, err)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(threatfox.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	tf := threatfox.New()
	gock.InterceptClient(tf.Client.HTTPClient)

	_, err = tf.Fetch()
	require.Error(t, err)
}
