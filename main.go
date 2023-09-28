package main

import (
	"errors"
	"github.com/bwmarrin/lit"
	"github.com/goccy/go-json"
	"github.com/kkyr/fig"
	"net/http"
	"strings"
	"time"
)

const baseAPIUrl = "https://api.cloudflare.com/client/v4/zones/"

var (
	records []zoneAndRecords
	cfg     config
)

func init() {
	err := fig.Load(&cfg, fig.File("config.yml"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	// Set lit.LogLevel to the given value
	switch strings.ToLower(cfg.LogLevel) {
	case "war", "warning":
		lit.LogLevel = lit.LogWarning

	case "info", "informational":
		lit.LogLevel = lit.LogInformational

	case "deb", "debug":
		lit.LogLevel = lit.LogDebug
	}

	// Create the file lastip if it doesn't exist
	if !fileExists("lastip") {
		writeIP("")
	}

	for {
		err = getRecords()
		if err != nil {
			lit.Info("Error getting records, retrying in 3 seconds")
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
}

func main() {
	var (
		ip    string
		newIP string
	)

	ip = readIP()

	for {
		newIP = getIP()

		if newIP != ip && newIP != "" {
			lit.Info("IP changed from " + ip + " to " + newIP)

			// If we don't get any errors, we save the new ip
			if updateDuckDNS(newIP) == nil && updateCloudflare(newIP) == nil {
				ip = newIP
				writeIP(newIP)
			}
		}

		time.Sleep(cfg.Timeout)
	}
}

// return A records hosted on cloudflare
func getRecords() error {
	request, err := http.NewRequest("GET", baseAPIUrl, nil)
	request.Header.Add("authorization", "Bearer "+cfg.Token)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.New("error while getting records: " + err.Error())
	}

	var zones apiZones
	_ = json.NewDecoder(response.Body).Decode(&zones)
	if zones.Success {
		for i, zone := range zones.Result {
			if allowedRecords, ok := cfg.Domains[zone.Name]; ok {
				records = append(records, zoneAndRecords{
					ZoneID:  zone.ID,
					Records: nil,
				})

				request, err = http.NewRequest("GET", baseAPIUrl+zone.ID+"/dns_records/", nil)
				request.Header.Add("authorization", "Bearer "+cfg.Token)

				response, err = http.DefaultClient.Do(request)
				if err != nil {
					return errors.New("error while getting records: " + err.Error())
				}

				var apiResponse apiRecords
				_ = json.NewDecoder(response.Body).Decode(&apiResponse)

				if apiResponse.Success {
					for _, record := range apiResponse.Result {
						if record.Type == "A" {
							if _, ok := allowedRecords.V4Records[record.Name]; ok {
								records[i].Records = append(records[i].Records, record)
							}
						}
					}
				}

			}
		}
	}

	return nil
}
