package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	listSep       = ','
	listSepString = string(listSep)
	typeSep       = '='
)

func toMapBool(src string) map[string]bool {
	m := map[string]bool{}
	if src != "" {
		for _, s := range strings.Split(src, listSepString) {
			m[s] = true
		}
	}
	return m
}

func toMapString(src string) (map[string]string, error) {
	m := map[string]string{}
	if src == "" {
		return m, nil
	}

	for _, kv := range strings.Split(src, listSepString) {
		i := strings.IndexByte(kv, typeSep)
		if i < 0 {
			return nil, fmt.Errorf("missing separator %c in %s", typeSep, kv)
		}
		m[kv[:i]] = kv[i+1:]
	}

	return m, nil
}

func toMapInt(src string) (map[string]int, error) {
	m := map[string]int{}
	if src == "" {
		return m, nil
	}

	for _, kv := range strings.Split(src, listSepString) {
		i := strings.IndexByte(kv, typeSep)
		if i < 0 {
			return nil, fmt.Errorf("missing separator %c in %s", typeSep, kv)
		}
		n, err := strconv.Atoi(kv[i+1:])
		if err != nil {
			return nil, err
		}
		m[kv[:i]] = n
	}

	return m, nil
}
