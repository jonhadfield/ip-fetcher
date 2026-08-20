// Package iplist parses the plain text address lists published by several
// providers, where each line holds a bare IP address or a CIDR prefix.
package iplist

import (
	"bufio"
	"bytes"
	"net/netip"

	"github.com/sirupsen/logrus"
)

// commentPrefix marks a line, or the remainder of a line, as a comment.
const commentPrefix = '#'

// ToPrefix accepts either a CIDR prefix or a bare IP address and returns a
// netip.Prefix, using a host prefix for bare addresses.
func ToPrefix(entry string) (netip.Prefix, bool) {
	if prefix, err := netip.ParsePrefix(entry); err == nil {
		return prefix, true
	}

	if addr, err := netip.ParseAddr(entry); err == nil {
		return netip.PrefixFrom(addr, addr.BitLen()), true
	}

	return netip.Prefix{}, false
}

// Parse reads a newline separated list of addresses or prefixes, splitting them
// by address family. Blank lines and comments are ignored, and entries that
// cannot be parsed are logged and skipped so one malformed line does not
// discard the whole list. source names the provider in those log messages.
//
// An error is only returned if the underlying data cannot be scanned.
func Parse(source string, data []byte) ([]netip.Prefix, []netip.Prefix, error) {
	var ipv4, ipv6 []netip.Prefix

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), bufio.MaxScanTokenSize)

	for scanner.Scan() {
		entry := strip(scanner.Bytes())
		if entry == "" {
			continue
		}

		prefix, ok := ToPrefix(entry)
		if !ok {
			logrus.Warnf("failed to parse %s address: %s", source, entry)

			continue
		}

		if prefix.Addr().Is4() {
			ipv4 = append(ipv4, prefix)

			continue
		}

		ipv6 = append(ipv6, prefix)
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return ipv4, ipv6, nil
}

// strip removes any trailing comment and surrounding whitespace from a line.
func strip(line []byte) string {
	if i := bytes.IndexByte(line, commentPrefix); i >= 0 {
		line = line[:i]
	}

	return string(bytes.TrimSpace(line))
}
