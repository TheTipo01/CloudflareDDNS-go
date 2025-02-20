package main

import (
	"github.com/bwmarrin/lit"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	replacer *strings.Replacer
	req      *http.Request
)

func init() {
	replacer = strings.NewReplacer("[", "{", "]", "}", "{", "", "}", "")
	req, _ = http.NewRequest("GET", cfg.Endpoint, nil)
	// Add Accept-Language header, otherwise the modem will throw bad requests at us
	req.Header.Set("Accept-Language", "it-IT")
}

func getIP() string {
	var out StationJSON

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lit.Error("Error while requesting ip: " + err.Error())
		return ""
	}

	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	// The JSON is given to us in an array. We parse that and remove the brackets, and add them only at the end
	_ = json.Unmarshal([]byte(replacer.Replace(string(b))), &out)

	return out.WanIP4Addr
}

// Overwrite the file lastip with the given IP
func writeIP(ip string) {
	err := os.WriteFile("lastip", []byte(ip), 0644)
	if err != nil {
		lit.Error("Error while writing lastip: " + err.Error())
	}
}

// Read the file lastip and return its content
func readIP() string {
	b, err := os.ReadFile("lastip")
	if err != nil {
		lit.Error("Error while reading lastip: " + err.Error())
		return ""
	}

	return string(b)
}

// Returns true if the given file exists
func fileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
