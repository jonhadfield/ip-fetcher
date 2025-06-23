package main

import (
	"fmt"
	"net/netip"
	"reflect"
	"strings"
)

// docToLines converts any provider document or prefix slice to newline separated IP prefixes.
func docToLines(doc any) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("no prefixes found")
	}

	var prefixes []string
	addrType := reflect.TypeOf(netip.Addr{})
	prefixType := reflect.TypeOf(netip.Prefix{})

	var walk func(val reflect.Value)
	walk = func(val reflect.Value) {
		if !val.IsValid() {
			return
		}
		if val.Type() == prefixType {
			prefixes = append(prefixes, val.Interface().(netip.Prefix).String())
			return
		}
		if val.Type() == addrType {
			prefixes = append(prefixes, val.Interface().(netip.Addr).String())
			return
		}
		switch val.Kind() {
		case reflect.Ptr:
			if !val.IsNil() {
				walk(val.Elem())
			}
		case reflect.Struct:
			for i := 0; i < val.NumField(); i++ {
				walk(val.Field(i))
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < val.Len(); i++ {
				walk(val.Index(i))
			}
		}
	}

	walk(reflect.ValueOf(doc))

	if len(prefixes) == 0 {
		return nil, fmt.Errorf("no prefixes found")
	}

	joined := strings.Join(prefixes, "\n") + "\n"

	return []byte(joined), nil
}
