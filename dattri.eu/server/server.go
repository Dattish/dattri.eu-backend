package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"dattri.eu/handler"
)

type config map[string]string

type certLocations struct {
	Cert    string `json:"cert"`
	Privkey string `json:"privkey"`
}

func getCertLocations(filename string) (certLocations, error) {
	raw, err := ioutil.ReadFile(filename)
	var certLocations certLocations
	json.Unmarshal(raw, &certLocations)

	return certLocations, err
}

func setUpEndpoints(config config, serveMux *http.ServeMux) {
	endpoints, IOError := ioutil.ReadFile(config["endpoints"])
	if IOError != nil {
		log.Fatal("Couldn't read endpoints file: ", IOError)
	}
	endpointsError := handler.EndpointsFromConfig(serveMux, endpoints)
	if endpointsError != nil {
		log.Fatal("Couldn't set up endpoints: ", endpointsError)
	}
}

func getCertFiles(config config) (certFile string, keyFile string) {
	certs, certErr := getCertLocations(config["certLocations"])
	if certErr != nil {
		log.Fatal("Couldn't get cert locations: ", certErr)
	}
	certFile = certs.Cert
	keyFile = certs.Privkey
	return
}

func main() {
	var config config
	configAsBytes, configIOError := ioutil.ReadFile("./config.json")
	if configIOError != nil {
		log.Fatal("Couldn't read config file: ", configIOError)
	}
	json.Unmarshal(configAsBytes, &config)

	serveMux := http.NewServeMux()

	setUpEndpoints(config, serveMux)

	serveMux.Handle("/monitoring", handler.Monitoring(time.Now()))

	certFile, keyFile := getCertFiles(config)

	fmt.Println("Running..")

	go http.ListenAndServe(config["httpPort"], handler.HttpsRedirect())
	err := http.ListenAndServeTLS(config["httpsPort"], certFile, keyFile,
			handler.CSP(config["contentPolicy"],
				handler.CORS(config["CORSMethods"], config["CORSOrigin"],
					handler.Logging(serveMux))))

	if err != nil {
		log.Fatal(err)
	}
}
