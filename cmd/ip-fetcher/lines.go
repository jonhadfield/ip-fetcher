package main

import (
	"errors"
	"fmt"
	"net/netip"
	"reflect"
	"strings"

	"github.com/jonhadfield/ip-fetcher/internal/aggregate"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/urfave/cli/v2"
)

var errNoPrefixes = errors.New("no prefixes found")

// prefixesToLines renders IPv4 then IPv6 prefixes as newline separated text.
func prefixesToLines(ipv4, ipv6 []netip.Prefix) []byte {
	sl := strings.Builder{}
	for x := range ipv4 {
		fmt.Fprintf(&sl, "%s\n", ipv4[x].String())
	}

	for x := range ipv6 {
		fmt.Fprintf(&sl, "%s\n", ipv6[x].String())
	}

	return []byte(sl.String())
}

// linesWanted reports whether the command is producing newline separated prefixes.
func linesWanted(c *cli.Context) bool {
	return c.Bool(formatLines) || c.String(flagFormat) == formatLines
}

// aggregateModeOrError parses --aggregate and ensures it is only used with lines output.
func aggregateModeOrError(c *cli.Context) (aggregate.Mode, error) {
	return aggregateMode(c, linesWanted(c))
}

// aggregateMode parses --aggregate. When lines is false, a non-None mode is an error.
func aggregateMode(c *cli.Context, lines bool) (aggregate.Mode, error) {
	mode, err := aggregate.ParseMode(c.String(flagAggregate))
	if err != nil {
		return aggregate.None, err
	}

	if mode != aggregate.None && !lines {
		return aggregate.None, errors.New(errAggregateNeedsLines)
	}

	return mode, nil
}

// aggregateFlag is the shared --aggregate exact|cover option.
func aggregateFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  flagAggregate,
		Usage: usageAggregate,
	}
}

// docToLinesWithCLI reads --aggregate from the CLI context and renders lines.
func docToLinesWithCLI(c *cli.Context, doc any) ([]byte, error) {
	mode, err := aggregateModeOrError(c)
	if err != nil {
		return nil, err
	}

	return docToLines(doc, mode)
}

// docToLines converts any provider document or prefix slice to newline separated
// IP prefixes, optionally aggregating them.
func docToLines(doc any, mode aggregate.Mode) ([]byte, error) {
	if doc == nil {
		return nil, errNoPrefixes
	}

	raw := collectPrefixes(doc)
	if len(raw) == 0 {
		return nil, errNoPrefixes
	}

	prefixes := make([]netip.Prefix, 0, len(raw))
	for _, s := range raw {
		p, ok := iplist.ToPrefix(s)
		if !ok {
			continue
		}

		prefixes = append(prefixes, p)
	}

	if len(prefixes) == 0 {
		return nil, errNoPrefixes
	}

	prefixes = aggregate.Apply(mode, prefixes)

	lines := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		lines = append(lines, p.String())
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

// aggregateLinesOutput re-parses newline separated prefixes and applies mode.
func aggregateLinesOutput(data []byte, mode aggregate.Mode) ([]byte, error) {
	if mode == aggregate.None {
		return data, nil
	}

	ipv4, ipv6, err := iplist.Parse("aggregate", data)
	if err != nil {
		return nil, err
	}

	if len(ipv4) == 0 && len(ipv6) == 0 {
		return nil, errNoPrefixes
	}

	return prefixesToLinesAggregated(ipv4, ipv6, mode), nil
}

// prefixesToLinesAggregated applies aggregation then renders IPv4 then IPv6.
func prefixesToLinesAggregated(ipv4, ipv6 []netip.Prefix, mode aggregate.Mode) []byte {
	merged := aggregate.Apply(mode, append(append([]netip.Prefix{}, ipv4...), ipv6...))

	var out4, out6 []netip.Prefix
	for _, p := range merged {
		if p.Addr().Is4() {
			out4 = append(out4, p)

			continue
		}

		out6 = append(out6, p)
	}

	return prefixesToLines(out4, out6)
}

func collectPrefixes(input any) []string { //nolint:gocognit
	prefixes := make([]string, 0)

	var walk func(value any)
	walk = func(value any) {
		switch typed := value.(type) {
		case netip.Prefix:
			prefixes = append(prefixes, typed.String())
			return
		case netip.Addr:
			prefixes = append(prefixes, typed.String())
			return
		case []netip.Prefix:
			for _, prefix := range typed {
				walk(prefix)
			}
			return
		case []netip.Addr:
			for _, addr := range typed {
				walk(addr)
			}
			return
		}

		rv := reflect.ValueOf(value)
		if !rv.IsValid() {
			return
		}

		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return
			}

			walk(rv.Elem().Interface())
			return
		}

		if rv.Kind() == reflect.Struct {
			rt := rv.Type()

			n := rv.NumField()
			for i := range n {
				// unexported fields (e.g. those within an embedded time.Time)
				// cannot be read via reflection, and hold no prefixes.
				if !rt.Field(i).IsExported() {
					continue
				}

				walk(rv.Field(i).Interface())
			}
			return
		}

		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			n := rv.Len()
			for i := range n {
				walk(rv.Index(i).Interface())
			}
		}
	}

	walk(input)

	return prefixes
}
