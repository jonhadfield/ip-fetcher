package iplist

import (
	"encoding/json"
	"net/netip"

	"github.com/sirupsen/logrus"
)

// Families is the combined representation of a provider whose address families
// arrive from separate endpoints, so both are stored in a single
// re-processable file.
type Families struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// MarshalFamilies renders the two lists as the document the provider publishes.
func MarshalFamilies(ipv4, ipv6 []string) ([]byte, error) {
	return json.MarshalIndent(Families{IPv4: ipv4, IPv6: ipv6}, "", " ")
}

// ParseFamilies reads that document back, splitting each family into prefixes.
// Entries may be bare addresses or prefixes; those that parse as neither are
// logged against source and skipped.
func ParseFamilies(source string, data []byte) ([]netip.Prefix, []netip.Prefix, error) {
	var families Families
	if err := json.Unmarshal(data, &families); err != nil {
		return nil, nil, err
	}

	return CastPrefixes(source, families.IPv4), CastPrefixes(source, families.IPv6), nil
}

// CastPrefixes turns published addresses into prefixes, using a host prefix for
// bare addresses and skipping anything that cannot be parsed.
func CastPrefixes(source string, in []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(in))

	for _, entry := range in {
		prefix, ok := ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse %s address: %s", source, entry)

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	if len(prefixes) == 0 {
		return nil
	}

	return prefixes
}
