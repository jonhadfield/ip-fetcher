package iplist_test

import (
	"net/netip"
	"testing"

	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/stretchr/testify/require"
)

// the combined document round trips, with each family kept apart.
func TestMarshalAndParseFamilies(t *testing.T) {
	data, err := iplist.MarshalFamilies([]string{"1.2.3.4", "10.0.0.0/8"}, []string{"2001:db8::1"})
	require.NoError(t, err)
	require.JSONEq(t, `{"ipv4":["1.2.3.4","10.0.0.0/8"],"ipv6":["2001:db8::1"]}`, string(data))

	ipv4, ipv6, err := iplist.ParseFamilies("test", data)
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, ipv4)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("2001:db8::1/128")}, ipv6)
}

func TestParseFamiliesInvalidJSON(t *testing.T) {
	_, _, err := iplist.ParseFamilies("test", []byte("not json"))
	require.Error(t, err)
}

// an unparseable entry is skipped rather than failing the whole list, and a
// family with nothing usable in it comes back empty.
func TestCastPrefixesSkipsInvalidEntries(t *testing.T) {
	require.Equal(
		t,
		[]netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")},
		iplist.CastPrefixes("test", []string{"1.2.3.4", "not-an-ip"}),
	)
	require.Empty(t, iplist.CastPrefixes("test", []string{"not-an-ip"}))
	require.Empty(t, iplist.CastPrefixes("test", nil))
}
