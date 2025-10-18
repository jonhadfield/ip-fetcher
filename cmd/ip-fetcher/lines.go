package main

import (
	"errors"
	"net/netip"
	"reflect"
	"strings"
)

var errNoPrefixes = errors.New("no prefixes found")

// docToLines converts any provider document or prefix slice to newline separated IP prefixes.
func docToLines(doc any) ([]byte, error) {
	if doc == nil {
		return nil, errNoPrefixes
	}

	prefixes := collectPrefixes(doc)
	if len(prefixes) == 0 {
		return nil, errNoPrefixes
	}

	joined := strings.Join(prefixes, "\n") + "\n"

	return []byte(joined), nil
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
			n := rv.NumField()
			for i := range n {
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
