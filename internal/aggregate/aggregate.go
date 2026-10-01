// Package aggregate reduces a set of IP prefixes to a smaller equivalent or
// covering set.
package aggregate

import (
	"fmt"
	"net/netip"
	"slices"

	"go4.org/netipx"
)

// Mode selects how prefixes are reduced.
type Mode int

const (
	// None leaves prefixes unchanged.
	None Mode = iota
	// Exact merges only when adjacent same-length prefixes form a clean
	// supernet, so the address union is unchanged.
	Exact
	// Cover produces fewer prefixes by filling holes between nearby ranges.
	// The result may include addresses that were not in the input.
	Cover
)

// ParseMode maps CLI values to a Mode. An empty string means None.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "", "none":
		return None, nil
	case "exact":
		return Exact, nil
	case "cover":
		return Cover, nil
	default:
		return None, fmt.Errorf("unknown aggregate mode %q (want exact or cover)", s)
	}
}

// Apply runs the selected mode over prefixes. None returns a defensive copy.
func Apply(mode Mode, prefixes []netip.Prefix) []netip.Prefix {
	switch mode {
	case Exact:
		return ExactMerge(prefixes)
	case Cover:
		return CoverMerge(prefixes)
	default:
		return slices.Clone(prefixes)
	}
}

// ExactMerge merges only hole-free sibling pairs and drops prefixes already
// covered by a broader input prefix. IPv4 and IPv6 are handled separately.
func ExactMerge(prefixes []netip.Prefix) []netip.Prefix {
	ipv4, ipv6 := splitFamilies(prefixes)

	return append(exactFamily(ipv4), exactFamily(ipv6)...)
}

// CoverMerge fills holes between nearby ranges when the gap is no larger than
// the wider of the two ranges, then emits a minimal prefix list for each
// resulting span. IPv4 and IPv6 are handled separately.
func CoverMerge(prefixes []netip.Prefix) []netip.Prefix {
	ipv4, ipv6 := splitFamilies(prefixes)

	return append(coverFamily(ipv4), coverFamily(ipv6)...)
}

func splitFamilies(prefixes []netip.Prefix) (ipv4, ipv6 []netip.Prefix) {
	for _, p := range prefixes {
		if !p.IsValid() {
			continue
		}

		if p.Addr().Is4() {
			ipv4 = append(ipv4, p)

			continue
		}

		ipv6 = append(ipv6, p)
	}

	return ipv4, ipv6
}

func exactFamily(prefixes []netip.Prefix) []netip.Prefix {
	if len(prefixes) == 0 {
		return nil
	}

	normalized := removeCovered(prefixes)

	for {
		merged, changed := mergePass(normalized)
		normalized = merged
		if !changed {
			break
		}
	}

	return normalized
}

func coverFamily(prefixes []netip.Prefix) []netip.Prefix {
	if len(prefixes) == 0 {
		return nil
	}

	ranges := prefixesToRanges(removeCovered(prefixes))
	if len(ranges) == 0 {
		return nil
	}

	slices.SortFunc(ranges, compareRange)

	merged := []netipx.IPRange{ranges[0]}
	for _, r := range ranges[1:] {
		last := merged[len(merged)-1]
		if !last.IsValid() || !r.IsValid() {
			merged = append(merged, r)

			continue
		}

		if canFillGap(last, r) {
			spanning := netipx.IPRangeFrom(last.From(), r.To())
			if spanning.IsValid() {
				merged[len(merged)-1] = spanning

				continue
			}
		}

		merged = append(merged, r)
	}

	out := make([]netip.Prefix, 0, len(merged))
	for _, r := range merged {
		out = append(out, coveringPrefix(r.From(), r.To()))
	}

	return out
}

// coveringPrefix returns the smallest prefix that fully contains the inclusive
// address range [from, to].
func coveringPrefix(from, to netip.Addr) netip.Prefix {
	if !from.IsValid() || !to.IsValid() || from.BitLen() != to.BitLen() {
		return netip.Prefix{}
	}

	if from.Compare(to) > 0 {
		from, to = to, from
	}

	for bits := from.BitLen(); bits >= 0; bits-- {
		p := netip.PrefixFrom(from, bits).Masked()
		if p.Contains(to) {
			return p
		}
	}

	return netip.Prefix{}
}

func prefixesToRanges(prefixes []netip.Prefix) []netipx.IPRange {
	ranges := make([]netipx.IPRange, 0, len(prefixes))
	for _, p := range prefixes {
		r := netipx.RangeOfPrefix(p)
		if r.IsValid() {
			ranges = append(ranges, r)
		}
	}

	return ranges
}

// canFillGap reports whether two sorted, non-overlapping ranges should be
// joined, filling the addresses between them. Overlapping or adjacent ranges
// always merge. A gap is filled only when it is no larger than the wider range.
func canFillGap(left, right netipx.IPRange) bool {
	if left.Overlaps(right) {
		return true
	}

	// Adjacent ranges (no gap): right starts at left.To()+1.
	if next, ok := addrAdd(left.To(), 1); ok && next == right.From() {
		return true
	}

	gapStart, ok := addrAdd(left.To(), 1)
	if !ok || gapStart.Compare(right.From()) >= 0 {
		return false
	}

	gapEnd, ok := addrAdd(right.From(), -1)
	if !ok {
		return false
	}

	gap := netipx.IPRangeFrom(gapStart, gapEnd)
	if !gap.IsValid() {
		return false
	}

	leftSize := rangeSize(left)
	rightSize := rangeSize(right)
	gapSize := rangeSize(gap)

	return gapSize <= max(leftSize, rightSize)
}

func rangeSize(r netipx.IPRange) uint64 {
	if !r.IsValid() {
		return 0
	}

	// Prefix-aligned ranges use bit length; otherwise approximate via prefix cover.
	prefixes := r.Prefixes()
	var n uint64
	for _, p := range prefixes {
		n += prefixSize(p)
	}

	return n
}

func prefixSize(p netip.Prefix) uint64 {
	bits := p.Addr().BitLen() - p.Bits()
	if bits >= 64 {
		// Cap so comparisons stay useful for huge IPv6 spans.
		return ^uint64(0)
	}

	return 1 << bits
}

func addrAdd(addr netip.Addr, delta int64) (netip.Addr, bool) {
	a := addr.As16()
	var carry int64 = delta
	for i := 15; i >= 0 && carry != 0; i-- {
		v := int64(a[i]) + carry
		carry = 0
		for v < 0 {
			v += 256
			carry--
		}
		for v > 255 {
			v -= 256
			carry++
		}
		a[i] = byte(v)
	}

	if carry != 0 {
		return netip.Addr{}, false
	}

	return netip.AddrFrom16(a).Unmap(), true
}

func compareRange(a, b netipx.IPRange) int {
	if c := a.From().Compare(b.From()); c != 0 {
		return c
	}

	return a.To().Compare(b.To())
}

// removeCovered drops duplicates and any prefix wholly contained in another.
func removeCovered(prefixes []netip.Prefix) []netip.Prefix {
	sorted := slices.Clone(prefixes)
	slices.SortFunc(sorted, comparePrefix)

	out := make([]netip.Prefix, 0, len(sorted))
	for _, p := range sorted {
		p = p.Masked()
		if !p.IsValid() {
			continue
		}

		if len(out) > 0 && out[len(out)-1] == p {
			continue
		}

		covered := false
		for _, kept := range out {
			if prefixContains(kept, p) {
				covered = true

				break
			}
		}

		if covered {
			continue
		}

		out = append(out, p)
	}

	return out
}

func prefixContains(outer, inner netip.Prefix) bool {
	if outer.Bits() > inner.Bits() {
		return false
	}

	return outer.Masked() == netip.PrefixFrom(inner.Addr(), outer.Bits()).Masked()
}

func mergePass(prefixes []netip.Prefix) ([]netip.Prefix, bool) {
	if len(prefixes) < 2 {
		return prefixes, false
	}

	sorted := slices.Clone(prefixes)
	slices.SortFunc(sorted, comparePrefix)

	out := make([]netip.Prefix, 0, len(sorted))
	changed := false

	for i := 0; i < len(sorted); {
		if i+1 < len(sorted) {
			if parent, ok := siblingParent(sorted[i], sorted[i+1]); ok {
				out = append(out, parent)
				changed = true
				i += 2

				continue
			}
		}

		out = append(out, sorted[i])
		i++
	}

	return out, changed
}

// siblingParent reports the /n-1 parent when a and b are the two halves of that
// parent (same length, adjacent, no hole).
func siblingParent(a, b netip.Prefix) (netip.Prefix, bool) {
	a = a.Masked()
	b = b.Masked()

	if a.Bits() != b.Bits() || a.Bits() == 0 || a.Addr().BitLen() != b.Addr().BitLen() {
		return netip.Prefix{}, false
	}

	parentBits := a.Bits() - 1
	parentA := netip.PrefixFrom(a.Addr(), parentBits).Masked()
	parentB := netip.PrefixFrom(b.Addr(), parentBits).Masked()
	if parentA != parentB || a == b {
		return netip.Prefix{}, false
	}

	return parentA, true
}

func comparePrefix(a, b netip.Prefix) int {
	if c := a.Addr().Compare(b.Addr()); c != 0 {
		return c
	}

	switch {
	case a.Bits() < b.Bits():
		return -1
	case a.Bits() > b.Bits():
		return 1
	default:
		return 0
	}
}
