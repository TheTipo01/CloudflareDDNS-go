package main

import (
	"errors"
	"github.com/goccy/go-json"
	"net/http"
	"strings"
)

// Updates DuckDNS
func updateDuckDNS(ip string) error {
	_, err := http.Get("https://www.duckdns.org/update?domains=" + cfg.DDDomain + "&token=" + cfg.DDToken + "&ip=" + ip)
	if err != nil {
		return errors.New("error while updating DuckDNS: " + err.Error())
	}

	return nil
}

func updateCloudflare(ip string) error {
	var err error
	for _, zone := range records {
		for _, record := range zone.Records {
			if record.Type == "A" {
				err = patchRecord(zone, record, ip)
			}
		}
	}

	return err
}

func patchRecord(zone zoneAndRecords, record dnsRecord, ip string) error {
	request, err := http.NewRequest("PATCH", baseAPIUrl+zone.ZoneID+"/dns_records/"+record.ID,
		strings.NewReader("{\"content\":\""+ip+"\"}"))
	request.Header.Add("authorization", "Bearer "+cfg.Token)
	request.Header.Add("content-type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.New("error while updating Cloudflare: " + err.Error())
	}

	var apiResponse apiFeedback
	_ = json.NewDecoder(response.Body).Decode(&apiResponse)
	if !apiResponse.Success {
		return errors.New("error while updating Cloudflare: " + apiResponse.Errors[0].Message)
	}

	return nil
}
