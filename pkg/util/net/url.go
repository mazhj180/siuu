package net

import (
	"net"
	"net/url"
	"strconv"
)

type ParsedURL struct {
	Scheme string
	Host   string
	Port   uint16
	URI    string
	Params map[string][]string
}

func ParseURL(rawURL string) (*ParsedURL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		host = parsed.Host
		port = "0"
	}

	portUint, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}

	query := parsed.Query()
	params := make(map[string][]string)
	for key, values := range query {
		params[key] = values
	}

	return &ParsedURL{
		Scheme: parsed.Scheme,
		Host:   host,
		Port:   uint16(portUint),
		URI:    parsed.Path,
		Params: params,
	}, nil
}
