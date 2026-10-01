package aggregate_test

import (
	"net/netip"
	"testing"

	"github.com/jonhadfield/ip-fetcher/internal/aggregate"
	"github.com/stretchr/testify/require"
)

func p(s string) netip.Prefix {
	return netip.MustParsePrefix(s)
}

func TestParseMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		want    aggregate.Mode
		wantErr bool
	}{
		{"", aggregate.None, false},
		{"none", aggregate.None, false},
		{"exact", aggregate.Exact, false},
		{"cover", aggregate.Cover, false},
		{"bogus", aggregate.None, true},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			got, err := aggregate.ParseMode(tc.in)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestExactMergesSiblings(t *testing.T) {
	t.Parallel()

	got := aggregate.ExactMerge([]netip.Prefix{
		p("10.0.0.0/25"),
		p("10.0.0.128/25"),
	})
	require.Equal(t, []netip.Prefix{p("10.0.0.0/24")}, got)
}

func TestExactKeepsNonAdjacent(t *testing.T) {
	t.Parallel()

	got := aggregate.ExactMerge([]netip.Prefix{
		p("10.0.0.0/25"),
		p("10.0.1.0/25"),
	})
	require.Equal(t, []netip.Prefix{
		p("10.0.0.0/25"),
		p("10.0.1.0/25"),
	}, got)
}

func TestCoverFillsHole(t *testing.T) {
	t.Parallel()

	got := aggregate.CoverMerge([]netip.Prefix{
		p("10.0.0.0/25"),
		p("10.0.1.0/25"),
	})
	require.Equal(t, []netip.Prefix{p("10.0.0.0/23")}, got)
}

func TestExactDropsCoveredChild(t *testing.T) {
	t.Parallel()

	got := aggregate.ExactMerge([]netip.Prefix{
		p("10.0.0.0/24"),
		p("10.0.0.0/25"),
	})
	require.Equal(t, []netip.Prefix{p("10.0.0.0/24")}, got)
}

func TestExactMixedFamilies(t *testing.T) {
	t.Parallel()

	got := aggregate.ExactMerge([]netip.Prefix{
		p("10.0.0.0/25"),
		p("10.0.0.128/25"),
		p("2001:db8::/33"),
		p("2001:db8:8000::/33"),
	})
	require.Equal(t, []netip.Prefix{
		p("10.0.0.0/24"),
		p("2001:db8::/32"),
	}, got)
}

func TestApplyNoneClones(t *testing.T) {
	t.Parallel()

	in := []netip.Prefix{p("10.0.0.0/25"), p("10.0.0.128/25")}
	got := aggregate.Apply(aggregate.None, in)
	require.Equal(t, in, got)
	got[0] = p("1.1.1.1/32")
	require.Equal(t, p("10.0.0.0/25"), in[0])
}

func TestExactRecursiveMerge(t *testing.T) {
	t.Parallel()

	got := aggregate.ExactMerge([]netip.Prefix{
		p("10.0.0.0/26"),
		p("10.0.0.64/26"),
		p("10.0.0.128/26"),
		p("10.0.0.192/26"),
	})
	require.Equal(t, []netip.Prefix{p("10.0.0.0/24")}, got)
}
