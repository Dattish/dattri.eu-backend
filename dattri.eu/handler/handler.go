package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

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

func proxyHandler(serveMux *http.ServeMux, endpoint string, method string, path string) {
	//TODO: make this testable, how to properly inject client?
	serveMux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			b, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			body = b
		}

		queryParams := r.URL.RawQuery
		if queryParams != ""{
			queryParams = "?" + queryParams
		}

		request, err := http.NewRequest(method, path + queryParams, bytes.NewBuffer(body))
		request.Header.Set("Content-Type", r.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		responseBody := &bytes.Buffer{}
		if response.Body != nil {
			_, err = responseBody.ReadFrom(response.Body)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer response.Body.Close()
		}

		w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
		w.Write(responseBody.Bytes())
	})
}

//Sets up endpoints for files and directories based on a json in []byte format
//Note that directory endpoints needs to begin and end with a slash, i.e "/test/"
func EndpointsFromConfig(serveMux *http.ServeMux, rawJson []byte) error {
	var endpoints []map[string]string
	json.Unmarshal(rawJson, &endpoints)

	for _, endpoint := range endpoints {
		switch resourceType := endpoint["resourceType"]; resourceType {
		case "file":
			err := fileHandler(serveMux, endpoint["endpoint"], endpoint["path"])
			if err != nil {
				return err
			}
		case "directory":
			err := directoryHandler(serveMux, endpoint["endpoint"], endpoint["path"])
			if err != nil {
				return err
			}
		case "proxy":
			proxyHandler(serveMux, endpoint["endpoint"], endpoint["method"], endpoint["path"])
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
