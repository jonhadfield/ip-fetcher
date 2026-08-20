package iplist_test

import (
	"net/netip"
	"testing"

	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/stretchr/testify/require"
)

func TestParseSplitsByFamily(t *testing.T) {
	ipv4, ipv6, err := iplist.Parse("test", []byte("1.2.3.4\n10.0.0.0/8\n2001:db8::1\n2001:db8::/32\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
	}, ipv4)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("2001:db8::1/128"),
		netip.MustParsePrefix("2001:db8::/32"),
	}, ipv6)
}

// whole line comments, trailing comments and blank lines carry no addresses.
func TestParseIgnoresCommentsAndBlanks(t *testing.T) {
	ipv4, ipv6, err := iplist.Parse("test", []byte("# header\n\n   \n1.2.3.4 # trailing\n#1.2.3.5\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/32")}, ipv4)
	require.Empty(t, ipv6)
}

// a malformed entry is skipped rather than failing the whole list.
func TestParseSkipsInvalidEntry(t *testing.T) {
	ipv4, ipv6, err := iplist.Parse("test", []byte("1.2.3.4\nnot-an-ip\n999.1.1.1\n5.6.7.8\n"))
	require.NoError(t, err)
	require.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
		netip.MustParsePrefix("5.6.7.8/32"),
	}, ipv4)
	require.Empty(t, ipv6)
}

func TestParseEmpty(t *testing.T) {
	ipv4, ipv6, err := iplist.Parse("test", nil)
	require.NoError(t, err)
	require.Empty(t, ipv4)
	require.Empty(t, ipv6)
}

func TestToPrefix(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"1.2.3.4", "1.2.3.4/32", true},
		{"10.0.0.0/8", "10.0.0.0/8", true},
		{"2001:db8::1", "2001:db8::1/128", true},
		{"2001:db8::/32", "2001:db8::/32", true},
		{"nope", "", false},
		{"", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, ok := iplist.ToPrefix(tc.in)
			require.Equal(t, tc.ok, ok)

			if tc.ok {
				require.Equal(t, netip.MustParsePrefix(tc.want), got)
			}
		})
	}
}

func TestToLines(t *testing.T) {
	got := iplist.ToLines([]netip.Prefix{
		netip.MustParsePrefix("2.16.0.0/13"),
		netip.MustParsePrefix("2a02:26f0::/32"),
	})
	require.Equal(t, "2.16.0.0/13\n2a02:26f0::/32", string(got))
}

func TestToLinesEmpty(t *testing.T) {
	require.Empty(t, iplist.ToLines(nil))
}
