package main

import (
	"time"
)

type config struct {
	Token    string        `fig:"token" validate:"required"`
	Timeout  time.Duration `fig:"timeout" default:"60s"`
	LogLevel string        `fig:"loglevel" default:"error"`
	Domains  map[string]struct {
		V4Records map[string]interface{} `fig:"v4-records"`
	} `fig:"zones" validate:"required"`
	Endpoint string `fig:"endpoint" validate:"required"`
	DDDomain string `fig:"dd_domain" validate:"required"`
	DDToken  string `fig:"dd_token" validate:"required"`
}

// StationJSON is the dumb structure used to unmarshall the request from the router
type StationJSON struct {
	WanIP4Addr string `json:"wan_ip4_addr"`
}

type apiZones struct {
	Result []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
	Success bool `json:"success"`
}

type apiRecords struct {
	Result  []dnsRecord `json:"result"`
	Success bool        `json:"success"`
	Errors  []apiError  `json:"errors"`
}

type apiFeedback struct {
	Success bool       `json:"success"`
	Errors  []apiError `json:"errors"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type dnsRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type zoneAndRecords struct {
	ZoneID  string
	Records []dnsRecord
}
