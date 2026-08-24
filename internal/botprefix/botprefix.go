// Package botprefix parses the crawler prefix document that Google publishes
// for its bots and that several other operators have adopted: a creationTime
// and a prefixes array whose entries carry either an ipv4Prefix or an
// ipv6Prefix.
package botprefix

import (
	"encoding/json"
	"net/netip"
	"time"
)

// DefaultTimeFormat is the creationTime layout these feeds use.
const DefaultTimeFormat = "2006-01-02T15:04:05.999999"

type RawIPv4Entry struct {
	IPv4Prefix string `json:"ipv4Prefix"`
}

type RawIPv6Entry struct {
	IPv6Prefix string `json:"ipv6Prefix"`
}

type RawDoc struct {
	CreationTime  string            `json:"creationTime"`
	LastRequested time.Time         `json:"-" yaml:"-"`
	Entries       []json.RawMessage `json:"prefixes"`
}

type IPv4Entry struct {
	IPv4Prefix netip.Prefix `json:"ipv4Prefix"`
}

type IPv6Entry struct {
	IPv6Prefix netip.Prefix `json:"ipv6Prefix"`
}

type Doc struct {
	CreationTime time.Time   `json:"creationTime" yaml:"creationTime"`
	IPv4Prefixes []IPv4Entry `json:"ipv4Prefixes" yaml:"ipv4Prefixes"`
	IPv6Prefixes []IPv6Entry `json:"ipv6Prefixes" yaml:"ipv6Prefixes"`
}

// Options covers the only way these feeds differ: how they treat creationTime.
type Options struct {
	// TimeFormats are tried in order. Empty means DefaultTimeFormat.
	TimeFormats []string

	// RequireCreationTime makes an absent creationTime an error. Google's feeds
	// always carry one; several of the others omit it.
	RequireCreationTime bool
}

func Parse(data []byte, opts Options) (Doc, error) {
	var rawDoc RawDoc

	if err := json.Unmarshal(data, &rawDoc); err != nil {
		return Doc{}, err
	}

	ipv4, ipv6, err := castEntries(rawDoc.Entries)
	if err != nil {
		return Doc{}, err
	}

	doc := Doc{IPv4Prefixes: ipv4, IPv6Prefixes: ipv6}

	if rawDoc.CreationTime == "" && !opts.RequireCreationTime {
		return doc, nil
	}

	creationTime, err := parseCreationTime(rawDoc.CreationTime, opts.TimeFormats)
	if err != nil {
		return Doc{}, err
	}

	doc.CreationTime = creationTime

	return doc, nil
}

// parseCreationTime tries each format in turn, reporting the last failure.
func parseCreationTime(in string, formats []string) (time.Time, error) {
	if len(formats) == 0 {
		formats = []string{DefaultTimeFormat}
	}

	var lastErr error

	for _, format := range formats {
		creationTime, err := time.Parse(format, in)
		if err == nil {
			return creationTime, nil
		}

		lastErr = err
	}

	return time.Time{}, lastErr
}

// castEntries sorts the prefix entries by address family. An entry carries
// either an ipv4Prefix or an ipv6Prefix, so each is tried in turn.
func castEntries(prefixes []json.RawMessage) ([]IPv4Entry, []IPv6Entry, error) {
	var (
		ipv4 []IPv4Entry
		ipv6 []IPv6Entry
	)

	for _, pr := range prefixes {
		var (
			ipv4entry RawIPv4Entry
			ipv6entry RawIPv6Entry
		)

		// try 4
		if err := json.Unmarshal(pr, &ipv4entry); err == nil {
			ipv4Prefix, parseError := netip.ParsePrefix(ipv4entry.IPv4Prefix)
			if parseError == nil {
				ipv4 = append(ipv4, IPv4Entry{IPv4Prefix: ipv4Prefix})

				continue
			}
		}

		// try 6
		ipv6Err := json.Unmarshal(pr, &ipv6entry)
		if ipv6Err == nil {
			ipv6Prefix, parseError := netip.ParsePrefix(ipv6entry.IPv6Prefix)
			if parseError != nil {
				return ipv4, ipv6, parseError
			}

			ipv6 = append(ipv6, IPv6Entry{IPv6Prefix: ipv6Prefix})

			continue
		}

		return ipv4, ipv6, ipv6Err
	}

	return ipv4, ipv6, nil
}
