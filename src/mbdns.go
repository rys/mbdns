package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type record struct {
	Domain string
	Token  string
	Host   string
	TTL    string
	Record string
}

const mbConfigFile string = "mbdns.conf"
const mbURLv4 string = "https://dnsapi4.mythic-beasts.com/"
const mbURLv6 string = "https://dnsapi6.mythic-beasts.com/"
const mbCommand string = "REPLACE %s %s %s DYNAMIC_IP"
const mbLoopWaitSeconds string = "300s"
const mbRecordUpdateWaitSeconds string = "1s"
const mbResponseError string = "updating %s.%s (%s, %s) failed with %s"
const mbResponseSuccess string = "updating %s.%s (%s, %s) succeeded"
const mbLogActivity string = "running %s for %s.%s"
const mbStartupBanner string = "mbdns %s %s (git %s)"
const mbConfigPathBanner string = "config path: %s"
const mbRecordA = "A"
const mbRecordAAAA = "AAAA"

var records []record

// BuildVersion passed in via ldflags
var BuildVersion string

// BuildDate passed in via ldflags
var BuildDate string

// GitRev passed in via ldflags
var GitRev string

func process() {
	loopSleepDuration, _ := time.ParseDuration(mbLoopWaitSeconds)
	recordSleepDuration, _ := time.ParseDuration(mbRecordUpdateWaitSeconds)

	for {
		for i := range records {
			command := fmt.Sprintf(mbCommand, records[i].Host, records[i].TTL, records[i].Record)
			logActivityMsg := fmt.Sprintf(mbLogActivity, command, records[i].Host, records[i].Domain)

			log.Println(logActivityMsg)

			mbURL := mbURLv4

			if records[i].Record != mbRecordA && records[i].Record != mbRecordAAAA {
				continue
			}

			if records[i].Record == mbRecordAAAA {
				mbURL = mbURLv6
			}

			response, err := http.PostForm(mbURL, url.Values{"domain": {records[i].Domain}, "password": {records[i].Token}, "command": {command}})

			if err != nil {
				log.Println(fmt.Sprintf(mbResponseError, records[i].Host, records[i].Domain, records[i].Record, records[i].TTL, err.Error()))
				continue
			}

			defer response.Body.Close()

			if response.StatusCode != 200 {
				log.Println(fmt.Sprintf(mbResponseError, records[i].Host, records[i].Domain, records[i].Record, records[i].TTL, response.Status))

				body, _ := ioutil.ReadAll(response.Body)
				log.Printf("%s", body)

				continue
			}

			log.Println(fmt.Sprintf(mbResponseSuccess, records[i].Host, records[i].Domain, records[i].Record, records[i].TTL))

			time.Sleep(recordSleepDuration)
		}

		time.Sleep(loopSleepDuration)
	}
}

func main() {
	log.SetOutput(os.Stdout)

	var configFile string
	var printVer bool
	flag.StringVar(&configFile, "config", mbConfigFile, "Config file path")
	flag.BoolVar(&printVer, "version", false, "Print version banner and exit")
	flag.Parse()

	if printVer {
		fmt.Println(fmt.Sprintf(mbStartupBanner, BuildVersion, BuildDate, GitRev))
		os.Exit(0)
	}

	log.Println(fmt.Sprintf(mbStartupBanner, BuildVersion, BuildDate, GitRev))

	log.Println(fmt.Sprintf(mbConfigPathBanner, configFile))

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatal("config does not exist. Exiting...")
	}

	f, err := os.Lstat(configFile)

	if err != nil {
		log.Fatal("could not stat config. Exiting...")
	}

	if f.Mode() != 0400 {
		log.Fatal("config is potentially insecure. Exiting...")
	}

	log.Println("mbdns reading config")

	tuples, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal("could not read config records. Exiting...")
	}

	err = json.Unmarshal(tuples, &records)

	if err != nil {
		log.Fatal("could not process config. Invalid JSON? Exiting...")
	}

	log.Println("mbdns is processing records")

	go process()
	select {}
}
