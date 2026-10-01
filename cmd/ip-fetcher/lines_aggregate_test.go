package main

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/jonhadfield/ip-fetcher/internal/aggregate"
	"github.com/stretchr/testify/require"
)

func TestDocToLinesExactAndCover(t *testing.T) {
	t.Parallel()

	doc := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/25"),
		netip.MustParsePrefix("10.0.0.128/25"),
		netip.MustParsePrefix("10.0.1.0/25"),
	}

	exact, err := docToLines(doc, aggregate.Exact)
	require.NoError(t, err)
	require.Equal(t, "10.0.0.0/24\n10.0.1.0/25\n", string(exact))

	cover, err := docToLines(doc, aggregate.Cover)
	require.NoError(t, err)
	require.Equal(t, "10.0.0.0/23\n", string(cover))
}

func TestDocToLinesExactReducesSiblings(t *testing.T) {
	t.Parallel()

	doc := []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/25"),
		netip.MustParsePrefix("192.168.0.128/25"),
	}

	out, err := docToLines(doc, aggregate.Exact)
	require.NoError(t, err)
	require.Equal(t, []string{"192.168.0.0/24"}, strings.Split(strings.TrimSpace(string(out)), "\n"))
}
