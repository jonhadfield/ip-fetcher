package icloudpr_test

import (
	"net/netip"
	"testing"

	"github.com/jonhadfield/ip-fetcher/providers/icloudpr"

	"github.com/stretchr/testify/require"
)

// The package level helpers delegate to internal/geocsv. They are part of the
// published api, so a broken delegation would be silent without these.
func TestDelegatedHelpers(t *testing.T) {
	require.True(t, icloudpr.IsIPv4("1.2.3.4"))
	require.False(t, icloudpr.IsIPv4("2001:db8::1"))
	require.True(t, icloudpr.IsIPv6("2001:db8::1"))
	require.False(t, icloudpr.IsIPv6("1.2.3.4"))
}

func TestDelegatedParse(t *testing.T) {
	records, err := icloudpr.Parse([]byte("1.2.3.0/24,GB,GB-EN,London,\n"))
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, netip.MustParsePrefix("1.2.3.0/24"), records[0].Prefix)
	require.Equal(t, "London", records[0].City)
}

// a zero value client falls back to the package's DownloadURL rather than
// requesting an empty address.
func TestEmptyDownloadURLDefaults(t *testing.T) {
	a := icloudpr.ICloudPrivateRelay{}
	require.Empty(t, a.DownloadURL)

	// the request itself will fail without a client, but the default must be
	// applied before it is attempted.
	_, _, _, _ = a.FetchData()
	require.Equal(t, icloudpr.DownloadURL, a.DownloadURL)
}
