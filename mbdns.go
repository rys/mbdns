package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Simple type to help with record type disambiguation
type MbdnsRecordType string

const (
	mbRecordTypeA    MbdnsRecordType = "A"
	mbRecordTypeAAAA MbdnsRecordType = "AAAA"
)

// Type to help us unmarshal the JSON config from disk
type MbdnsConfig struct {
	Zone   string          `json:"zone"`
	Host   string          `json:"host"`
	KeyID  string          `json:"key_id"`
	Secret string          `json:"secret"`
	Type   MbdnsRecordType `json:"type"`
}

// Main error type and strings used when calling the HTTPS API endpoints
type MbdnsUpdateRecordError struct {
	Err error
}

func (m *MbdnsUpdateRecordError) Error() string {
	return fmt.Sprintf("update record error: %v", m.Err)
}

const (
	mbErrorClientConnection string = "ERROR_CLIENT_CONNECTION"
	mbErrorClientResponse   string = "ERROR_CLIENT_RESPONSE"
	mbErrorRequestCreation  string = "ERROR_REQUEST_CREATION"
	mbErrorUpdateFailure    string = "ERROR_UPDATE_FAILURE"
)

// Log strings printed when starting up
const (
	mbLogNoConfig            string = "config file does not exist"
	mbLogFailedConfigStat    string = "failed to stat config file"
	mbLogInsecureConfig      string = "config file is insecure"
	mbLogReadingConfig       string = "mbdns is reading its config"
	mbLogCannotReadConfig    string = "cannot read config file"
	mbLogCannotProcessConfig string = "cannot process config file"
	mbLogProcessingRecords   string = "mbdns is processing records"
)

// Help texts
const (
	mbHelpConfigFilePath      string = "Config file path"
	mbHelpPrintVersion        string = "Print version banner and exit"
	mbHelpAllowInsecureConfig string = "Allow reading an insecure config"
)

// Main application configuration and logging
const (
	mbConfigFile              string = "mbdns.conf"
	mbURLv4_2                 string = "https://ipv4.api.mythic-beasts.com/dns/v2/dynamic/%s.%s"
	mbURLv6_2                 string = "https://ipv6.api.mythic-beasts.com/dns/v2/dynamic/%s.%s"
	mbLoopWaitSeconds         string = "300s"
	mbRecordUpdateWaitSeconds string = "1s"
	mbStartupBanner           string = "mbdns %s %s (git %s) by %s"
	mbConfigPathBanner        string = "config path: %s"
	mbUpdateResponseError     string = "updating %s.%s (%s) failed with %s"
	mbUpdateResponseSuccess   string = "updating %s.%s (%s) succeeded"
)

// unmarshalled running configuration
var mbdnsConfig []MbdnsConfig

// Build-time version information
var BuildVersion string
var BuildDate string
var GitRev string
var BuildUser string

// call the Mythic Beasts DNS v2 API endpoint for dynamic DNS
func updateDynamicDNSv2(url string, clientID string, clientSecret string) error {
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(""))

	if err != nil {
		return &MbdnsUpdateRecordError{Err: errors.New(mbErrorRequestCreation)}
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return &MbdnsUpdateRecordError{Err: errors.New(mbErrorUpdateFailure)}
	}

	if response.StatusCode != 200 {
		return &MbdnsUpdateRecordError{Err: errors.New(mbErrorClientResponse + " code: " + fmt.Sprint(response.StatusCode))}
	}

	return nil
}

// main processing loop that works on config records
func processRecords() {
	loopSleepDuration, _ := time.ParseDuration(mbLoopWaitSeconds)
	recordSleepDuration, _ := time.ParseDuration(mbRecordUpdateWaitSeconds)

	for {
		for i := range mbdnsConfig {
			mbURL := fmt.Sprintf(mbURLv4_2, mbdnsConfig[i].Host, mbdnsConfig[i].Zone)

			if mbdnsConfig[i].Type == mbRecordTypeAAAA {
				mbURL = fmt.Sprintf(mbURLv6_2, mbdnsConfig[i].Host, mbdnsConfig[i].Zone)
			}

			err := updateDynamicDNSv2(mbURL, mbdnsConfig[i].KeyID, mbdnsConfig[i].Secret)

			if err != nil {
				log.Printf(mbUpdateResponseError, mbdnsConfig[i].Host, mbdnsConfig[i].Zone, mbdnsConfig[i].Type, err.Error())
				continue
			}

			log.Printf(mbUpdateResponseSuccess, mbdnsConfig[i].Host, mbdnsConfig[i].Zone, mbdnsConfig[i].Type)

			time.Sleep(recordSleepDuration)
		}

		time.Sleep(loopSleepDuration)
	}
}

func main() {
	log.SetOutput(os.Stdout)

	var configFile string
	var printVer bool
	var insecureConf bool

	flag.StringVar(&configFile, "config", mbConfigFile, mbHelpConfigFilePath)
	flag.BoolVar(&printVer, "version", false, mbHelpPrintVersion)
	flag.BoolVar(&insecureConf, "insecure", false, mbHelpAllowInsecureConfig)
	flag.Parse()

	if printVer {
		fmt.Printf(mbStartupBanner, BuildVersion, BuildDate, GitRev, BuildUser)
		os.Exit(0)
	}

	log.Printf(mbStartupBanner, BuildVersion, BuildDate, GitRev, BuildUser)
	log.Printf(mbConfigPathBanner, configFile)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatal(mbLogNoConfig)
	}

	f, err := os.Lstat(configFile)

	if err != nil {
		log.Fatal(mbLogFailedConfigStat)
	}

	if f.Mode() != 0400 {
		if !insecureConf {
			log.Fatal(mbLogInsecureConfig)
		}
	}

	log.Println(mbLogReadingConfig)

	tuples, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal(mbLogCannotReadConfig)
	}

	err = json.Unmarshal(tuples, &mbdnsConfig)

	if err != nil {
		log.Fatal(mbLogCannotProcessConfig)
	}

	log.Println(mbLogProcessingRecords)

	go processRecords()
	select {}
}
