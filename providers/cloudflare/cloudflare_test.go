package cloudflare_test

import (
	"fmt"
	"net/netip"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestFetch4(t *testing.T) {
	u, err := url.Parse(DefaultIPv4URL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ips-v4")

	cf := New()
	gock.InterceptClient(cf.Client.HTTPClient)

	p4, err := cf.Fetch4()
	require.NoError(t, err)
	require.NotEmpty(t, p4)
	require.Contains(t, p4, netip.MustParsePrefix("162.158.0.0/16"))
}

func TestFetch6(t *testing.T) {
	u, err := url.Parse(DefaultIPv6URL)
	require.NoError(t, err)

	urlBase := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	gock.New(urlBase).
		Get(u.Path).
		Reply(200).
		File("testdata/ips-v6")

	cf := New()
	gock.InterceptClient(cf.Client.HTTPClient)

	p6, err := cf.Fetch6()
	require.NoError(t, err)
	require.NotEmpty(t, p6)
	require.Contains(t, p6, netip.MustParsePrefix("2606:4700::/32"))
}

func TestFetch(t *testing.T) {
	u6, err := url.Parse(DefaultIPv6URL)
	require.NoError(t, err)

	urlBase6 := fmt.Sprintf("%s://%s", u6.Scheme, u6.Host)

	gock.New(urlBase6).
		Get(u6.Path).
		Reply(200).
		File("testdata/ips-v6")

	u4, err := url.Parse(DefaultIPv4URL)
	require.NoError(t, err)

	urlBase4 := fmt.Sprintf("%s://%s", u4.Scheme, u4.Host)
	gock.New(urlBase4).
		Get(u4.Path).
		Reply(200).
		File("testdata/ips-v4")

	cf := New()
	gock.InterceptClient(cf.Client.HTTPClient)

	ips, err := cf.Fetch()
	require.NoError(t, err)
	require.Len(t, ips, 22)
	require.Contains(t, ips, netip.MustParsePrefix("2606:4700::/32"))
	require.Contains(t, ips, netip.MustParsePrefix("131.0.72.1/22"))
}
