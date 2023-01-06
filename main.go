package main

import (
	"encoding/json"
	"github.com/bwmarrin/lit"
	"github.com/kkyr/fig"
	"net/http"
	"strings"
	"sync"
	"time"
)

const baseAPIUrl = "https://api.cloudflare.com/client/v4/zones/"

var (
	records   []zoneAndRecords
	cfg       config
	errorFlag bool
	wg        sync.WaitGroup
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
		getRecords()
		if errorFlag {
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

		if newIP != ip {
			lit.Info("IP changed from " + ip + " to " + newIP)

			wg.Add(2)
			go updateDuckDNS(newIP)
			go updateCloudflare(newIP)
			wg.Wait()

			// If we don't get any errors, we save the new ip
			if !errorFlag {
				ip = newIP
				writeIP(newIP)
			} else {
				errorFlag = false
			}
		}

		time.Sleep(cfg.Timeout)
	}
}

// return A and AAAA records hosted on cloudflare
func getRecords() {
	request, err := http.NewRequest("GET", baseAPIUrl, nil)
	request.Header.Add("authorization", "Bearer "+cfg.Token)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		lit.Error("%s", err)
		errorFlag = true
		return
	}

	var zones apiZones
	_ = json.NewDecoder(response.Body).Decode(&zones)
	if zones.Success {
		wg := sync.WaitGroup{}
		defer wg.Wait()
		for i, zone := range zones.Result {
			if allowedRecords, ok := cfg.Domains[zone.Name]; ok {
				wg.Add(1)
				records = append(records, zoneAndRecords{
					ZoneID:  zone.ID,
					Records: nil,
				})
				v4mutex := sync.Mutex{}
				i, zone := i, zone
				go func() {
					defer wg.Done()
					request, err := http.NewRequest("GET", strings.Join([]string{baseAPIUrl, zone.ID, "/dns_records/"}, ""), nil)
					request.Header.Add("authorization", "Bearer "+cfg.Token)

					response, err := http.DefaultClient.Do(request)
					if err != nil {
						lit.Error("%s", err)
						errorFlag = true
						return
					}

					var apiResponse apiRecords
					_ = json.NewDecoder(response.Body).Decode(&apiResponse)

					if apiResponse.Success {
						for _, record := range apiResponse.Result {
							if record.Type == "A" {
								if _, ok := allowedRecords.V4Records[record.Name]; ok {
									v4mutex.Lock()
									records[i].Records = append(records[i].Records, record)
									v4mutex.Unlock()
								}
							}
						}
					}
				}()
			}
		}
	}

	errorFlag = false
}
