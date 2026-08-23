package botprefix_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/jonhadfield/ip-fetcher/internal/botprefix"

	"github.com/stretchr/testify/require"
)

const bothFamilies = `{"creationTime":"2026-08-01T10:00:00.000000","prefixes":[` +
	`{"ipv4Prefix":"54.36.148.0/23"},{"ipv6Prefix":"2001:db8::/32"}]}`

func TestParse(t *testing.T) {
	doc, err := botprefix.Parse([]byte(bothFamilies), botprefix.Options{})
	require.NoError(t, err)
	require.Equal(t, []botprefix.IPv4Entry{
		{IPv4Prefix: netip.MustParsePrefix("54.36.148.0/23")},
	}, doc.IPv4Prefixes)
	require.Equal(t, []botprefix.IPv6Entry{
		{IPv6Prefix: netip.MustParsePrefix("2001:db8::/32")},
	}, doc.IPv6Prefixes)
	require.Equal(t,
		time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC),
		doc.CreationTime)
}

// The three feeds differ only in how they treat creationTime, so each policy
// is pinned here.
func TestCreationTimePolicies(t *testing.T) {
	noTime := `{"prefixes":[{"ipv4Prefix":"1.2.3.0/24"}]}`

	t.Run("optional accepts a feed without creationTime", func(t *testing.T) {
		doc, err := botprefix.Parse([]byte(noTime), botprefix.Options{})
		require.NoError(t, err)
		require.True(t, doc.CreationTime.IsZero())
		require.Len(t, doc.IPv4Prefixes, 1)
	})

	t.Run("required rejects a feed without creationTime", func(t *testing.T) {
		_, err := botprefix.Parse([]byte(noTime), botprefix.Options{RequireCreationTime: true})
		require.Error(t, err)
	})

	t.Run("rfc3339 needs the format listed", func(t *testing.T) {
		rfc := `{"creationTime":"2026-08-01T10:00:00Z","prefixes":[]}`

		_, err := botprefix.Parse([]byte(rfc), botprefix.Options{})
		require.Error(t, err)

		doc, err := botprefix.Parse([]byte(rfc), botprefix.Options{
			TimeFormats: []string{time.RFC3339, botprefix.DefaultTimeFormat},
		})
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC), doc.CreationTime.UTC())
	})

	t.Run("later formats are tried after earlier ones fail", func(t *testing.T) {
		doc, err := botprefix.Parse([]byte(bothFamilies), botprefix.Options{
			TimeFormats: []string{time.RFC3339, botprefix.DefaultTimeFormat},
		})
		require.NoError(t, err)
		require.False(t, doc.CreationTime.IsZero())
	})
}

func TestParseEmptyPrefixes(t *testing.T) {
	doc, err := botprefix.Parse([]byte(`{"creationTime":"2026-08-01T10:00:00.000000","prefixes":[]}`), botprefix.Options{})
	require.NoError(t, err)
	require.Empty(t, doc.IPv4Prefixes)
	require.Empty(t, doc.IPv6Prefixes)
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := botprefix.Parse([]byte("not json"), botprefix.Options{})
	require.Error(t, err)
}

// an unparseable ipv6 prefix is reported rather than skipped.
func TestParseInvalidIPv6Prefix(t *testing.T) {
	_, err := botprefix.Parse([]byte(`{"prefixes":[{"ipv6Prefix":"nonsense"}]}`), botprefix.Options{})
	require.Error(t, err)
}

// an unparseable creationTime is an error even when it is optional, since the
// field is present.
func TestParseInvalidCreationTime(t *testing.T) {
	_, err := botprefix.Parse([]byte(`{"creationTime":"nonsense","prefixes":[]}`), botprefix.Options{})
	require.Error(t, err)
}
