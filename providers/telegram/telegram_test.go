package telegram_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/telegram"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(telegram.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusOK).
		File("testdata/telegram.txt")

	tg := telegram.New()
	gock.InterceptClient(tg.Client.HTTPClient)

	doc, err := tg.Fetch()
	require.NoError(t, err)
	require.Contains(t, doc.IPv4Prefixes, netip.MustParsePrefix("91.108.56.0/22"))
	require.Contains(t, doc.IPv6Prefixes, netip.MustParsePrefix("2001:b28:f23d::/48"))
}

func TestProcessData(t *testing.T) {
	data, err := os.ReadFile("testdata/telegram.txt")
	require.NoError(t, err)

	doc, err := telegram.ProcessData(data)
	require.NoError(t, err)
	require.NotEmpty(t, doc.IPv4Prefixes)
	require.NotEmpty(t, doc.IPv6Prefixes)
}

func TestFetchBadStatus(t *testing.T) {
	defer gock.Off()

	u, err := url.Parse(telegram.DownloadURL)
	require.NoError(t, err)

	gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
		Get(u.Path).
		Reply(http.StatusNotFound)

	tg := telegram.New()
	gock.InterceptClient(tg.Client.HTTPClient)

	_, err = tg.Fetch()
	require.Error(t, err)
}
