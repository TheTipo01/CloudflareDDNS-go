package main

import (
	"encoding/json"
	"github.com/bwmarrin/lit"
	"net/http"
	"strings"
)

// Updates DuckDNS
func updateDuckDNS(ip string) {
	_, err := http.Get("https://www.duckdns.org/update?domains=" + cfg.DDDomain + "&token=" + cfg.DDToken + "&ip=" + ip)
	if err != nil {
		lit.Error("Error while updating DuckDNS: " + err.Error())
		errorFlag = true
	}

	wg.Done()
}

func updateCloudflare(ip string) {
	for _, zone := range records {
		zone := zone
		for _, record := range zone.Records {
			record := record
			if record.Type == "A" {
				wg.Add(1)
				go patchRecord(zone, record, ip)
			}
		}
	}
	wg.Done()
}

func patchRecord(zone zoneAndRecords, record dnsRecord, ip string) {
	request, err := http.NewRequest("PATCH", strings.Join([]string{baseAPIUrl, zone.ZoneID, "/dns_records/", record.ID}, ""),
		strings.NewReader("{\"content\":\""+ip+"\"}"))
	request.Header.Add("authorization", "Bearer "+cfg.Token)
	request.Header.Add("content-type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		lit.Error("%s", err)
		errorFlag = true
		return
	}

	var apiResponse apiFeedback
	_ = json.NewDecoder(response.Body).Decode(&apiResponse)
	if !apiResponse.Success {
		lit.Error("\n%s", apiResponse.Errors)
	}

	wg.Done()
}
