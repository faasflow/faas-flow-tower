package main

import (
	"encoding/json"
	"fmt"
	pagegen "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	jsonType = "application/json"
	htmlType = "text/html"
)

var (
	public_uri  = ""
	gateway_url = ""

	gen *pagegen.Template = nil
)

type Message struct {
	Method   string `json:"method"`
	Function string `json:"function"`
}

type HtmlObject struct {
	PublicURL string
	Functions []*Function
}

type Function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Dag             string            `json:"dag,omitempty"`
}

// initialize globals
func initialize() error {
	public_uri = os.Getenv("gateway_public_uri")
	gateway_url = os.Getenv("gateway_url")
	gen = pagegen.Must(pagegen.ParseGlob("assets/*.html"))
	return nil
}

// handle request
func requestHandler(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("accept")

	// Check if file request
	files, ok := r.URL.Query()["file"]
	if ok && len(files[0]) > 0 {
		sendFile(w, r, files[0])
		return
	}

	// Check if UI request
	if strings.Contains(accept, "html") {
		pageHandle(w)
		return
	}

	// If API request
	if strings.Contains(accept, "json") {

		if r.Body == nil {
			http.Error(w, "", 500)
			return
		}

		var msg Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		method := msg.Method

		switch method {

		case "state":
			listRequestHandle(w)
			return

		case "flow":
			function := msg.Function
			flowFunctionRequestHandle(w, function)
			return
		}
	}

}

// API Requests
func listRequestHandle(w http.ResponseWriter) {

	w.Header().Set("Content-Type", jsonType)
	functions, err := listFunction()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle list request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(functions)
	w.Write(data)
}

// flowFunctionRequestHandle request handler for a flow function
func flowFunctionRequestHandle(w http.ResponseWriter, function string) {
	w.Header().Set("Content-Type", jsonType)
	functions, err := listFunction()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle request, error: %v", err), http.StatusInternalServerError)
		return
	}
	for _, functionObj := range functions {
		if functionObj.Name == function {
			dog, derr := getDag(function)
			if derr != nil {
				http.Error(w, fmt.Sprintf("failed to handle request, %v", derr), http.StatusInternalServerError)
				return
			}
			functionObj.Dag = dog
			data, _ := json.Marshal(functionObj)
			w.Write(data)
			return
		}
	}
	http.Error(w, fmt.Sprintf("failed to handle request, function not found"), http.StatusInternalServerError)
}

// listFunction request to list-flow-function to get flow-function list
func listFunction() ([]*Function, error) {
	var err error

	c := http.Client{}

	request, _ := http.NewRequest(http.MethodGet, gateway_url+"function/list-flow-functions", nil)
	response, err := c.Do(request)

	if err == nil {
		defer response.Body.Close()

		if response.Body != nil {
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				log.Fatal(bErr)
			}

			functions := []*Function{}
			mErr := json.Unmarshal(bodyBytes, &functions)
			if mErr != nil {
				return nil, fmt.Errorf("failed to get function list, %v", mErr)
			}

			return functions, nil
		}
		return make([]*Function, 0), nil
	}

	return nil, fmt.Errorf("failed to get function list, %v", err)
}

// getDag request to dot-generator for the dag dot graph
func getDag(function string) (string, error) {
	var err error

	c := http.Client{}

	request, _ := http.NewRequest(http.MethodGet, gateway_url+"function/dot-generator?function="+function, nil)
	response, err := c.Do(request)
	if err == nil {
		defer response.Body.Close()

		if response.Body != nil {
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				return "", fmt.Errorf("failed to get dag, %v", bErr)
			}
			return string(bodyBytes), nil
		}
		return "", fmt.Errorf("failed to get dag, empty reply")
	}
	return "", fmt.Errorf("failed to get dag, %v", err)
}

// Handle UI
func pageHandle(w http.ResponseWriter) {

	functions, err := listFunction()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}

	htmlObj := HtmlObject{PublicURL: public_uri, Functions: functions}

	err = gen.ExecuteTemplate(w, "index", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}
}

// Static file request handler
func sendFile(w http.ResponseWriter, r *http.Request, file string) {
	filepath := "./assets/" + file
	log.Printf("Serving file %s", filepath)
	http.ServeFile(w, r, filepath)
}

func main() {

	err := initialize()
	if err != nil {
		log.Fatal("failed to initialize the gateway, error: ", err.Error())
	}
	log.Printf("successfully initialized gateway")

	http.HandleFunc("/", requestHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
