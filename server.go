package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dattish/trieu.io-backend/handler"
)

type config map[string]interface{}

type certLocations struct {
	Cert    string `json:"cert"`
	Privkey string `json:"privkey"`
}

func getCertLocations(filename string) (certLocations, error) {
	raw, err := ioutil.ReadFile(filename)
	var certLocations certLocations
	_ = json.Unmarshal(raw, &certLocations)

	return certLocations, err
}

func listenOnChanges(endpointsFile string, notifier chan bool) {
	lastChanged := time.Time{}
	for {
		fileInfo, err := os.Stat(endpointsFile)
		if err != nil {
			log.Fatal("Couldn't read endpoints file: ", err)
		}

		changeTimestamp := fileInfo.ModTime()

		if changeTimestamp.After(lastChanged) {
			lastChanged = changeTimestamp
			notifier <- true
		}

		time.Sleep(5 * time.Second)
	}
}

func listenAndServeTLS(config config, certFile string, keyFile string) error {
	notifier := make(chan bool)
	conf, err := filepath.Abs(config["endpoints"].(string))
	if err != nil {
		return err
	}
	go listenOnChanges(conf, notifier)
	var server *http.Server
	monitoring := handler.Monitoring(time.Now().UTC())
	ping := handler.Ping()

	for range notifier {
		log.Println("Loading new config..")
		if server != nil {
			ctx, _ := context.WithTimeout(context.Background(), time.Duration(10 * time.Second))
			if err := server.Shutdown(ctx); err != nil {
				return err
			}
		}
		serveMux := http.NewServeMux()
		serveMux.Handle("/monitoring", monitoring)
		serveMux.Handle("/monitoring/ping", ping)

		endpoints, err := ioutil.ReadFile(config["endpoints"].(string))
		if err != nil {
			return fmt.Errorf("couldn't load endpoints file: %v", err)
		}
		err = handler.EndpointsFromConfig(serveMux, endpoints)
		if err != nil {
			return fmt.Errorf("couldn't set up endpoints: %v", err)
		}

		untypedGarbage := config["CORSExceptions"].([]interface{})
		corsExceptions := make([]string, len(untypedGarbage))
		for i, v := range untypedGarbage {
			corsExceptions[i] = v.(string)
		}

		server = &http.Server{Addr: config["httpsPort"].(string),
			Handler: handler.CSP(config["contentPolicy"].(string), corsExceptions,
				handler.CORS(config["CORSMethods"].(string), config["CORSOrigin"].(string), corsExceptions,
					handler.Logging(serveMux)))}
		go server.ListenAndServeTLS(certFile, keyFile)
	}
	return nil
}

func getCertFiles(config config) (certFile string, keyFile string) {
	certs, certErr := getCertLocations(config["certLocations"].(string))
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
	_ = json.Unmarshal(configAsBytes, &config)

	certFile, keyFile := getCertFiles(config)

	fmt.Println("Running..")
	go http.ListenAndServe(config["httpPort"].(string), handler.HttpsRedirect())
	if err := listenAndServeTLS(config, certFile, keyFile); err != nil {
		log.Fatal(err)
	}
}
