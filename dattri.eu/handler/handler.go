package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type endpoint struct {
	Endpoint     string `json:"endpoint"`
	ResourceType string `json:"resourceType"`
	Path         string `json:"path"`
}

type endpointError struct {
	error string
}

func (e endpointError) Error() string {
	return e.error
}

func fileHandler(serveMux *http.ServeMux, endpoint string, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &endpointError{err.Error()}
	}
	serveMux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	})
	return nil
}

func directoryHandler(serveMux *http.ServeMux, endpoint string, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &endpointError{err.Error()}
	}
	serveMux.Handle(endpoint, http.StripPrefix(endpoint, http.FileServer(http.Dir(path))))
	return nil
}

//Sets up endpoints for files and directories based on a json in []byte format
//Note that directory endpoints needs to begin and end with a slash, i.e "/test/"
func EndpointsFromConfig(serveMux *http.ServeMux, rawJson []byte) error {
	var endpoints []endpoint
	json.Unmarshal(rawJson, &endpoints)

	for _, endpoint := range endpoints {
		switch resourceType := endpoint.ResourceType; resourceType {
		case "file":
			err := fileHandler(serveMux, endpoint.Endpoint, endpoint.Path)
			if err != nil {
				return err
			}
		case "directory":
			err := directoryHandler(serveMux, endpoint.Endpoint, endpoint.Path)
			if err != nil {
				return err
			}
		default:
			return &endpointError{fmt.Sprintf("No such resourceType: %s", resourceType)}
		}
	}
	return nil
}

func HttpsRedirect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req,
			"https://"+req.Host+req.URL.String(),
			http.StatusMovedPermanently)
	})
}
