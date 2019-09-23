package main

import (
	"encoding/json"
	"fmt"
	pagegen "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

const (
	jsonType = "application/json"
	htmlType = "text/html"
)

var (
	public_uri                    = ""
	gateway_url                   = ""
	gen         *pagegen.Template = nil

	acceptingConnections int32
)

// Message API request query
type Message struct {
	Method   string `json:"method"`
	Function string `json:"function"`
	TraceID  string `json:"trace-id"`
}

// HtmlObject object to render webpages
type HtmlObject struct {
	PublicURL string
	Functions []*Function
	Flow      string
	Requests  map[string]string
}

// Function object to retrive and response flow-function details
type Function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Dag             string            `json:"dag,omitempty"`
}

// NodeTrace traces of each nodes in a dag
type NodeTrace struct {
	StartTime int `json:"start-time"`
	Duration  int `json:"duration"`
	// Other can be added based on the needs
}

// RequestTrace object to retrive and response traces details
type RequestTrace struct {
	RequestID  string                `json:"request-id"`
	NodeTraces map[string]*NodeTrace `json:"traces"`
	StartTime  int                   `json:"start-time"`
	Duration   int                   `json:"duration"`
}

// listFlowsApiHandler handle api request to list flow function
func listFlowsApiHandler(w http.ResponseWriter) {

	w.Header().Set("Content-Type", jsonType)
	functions, err := listFlowFunctions()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle list request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(functions, "", "    ")
	w.Write(data)
}

// flowDescApiHandler request handler for a flow function
func flowDescApiHandler(w http.ResponseWriter, function string) {
	w.Header().Set("Content-Type", jsonType)
	functions, err := listFlowFunctions()
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
			data, _ := json.MarshalIndent(functionObj, "", "    ")
			w.Write(data)
			return
		}
	}
	http.Error(w, fmt.Sprintf("failed to handle request, function not found"), http.StatusInternalServerError)
}

// listFlowRequestsApiHandler list the requests for a flow function
func listFlowRequestsApiHandler(w http.ResponseWriter, flowfunction string) {
	w.Header().Set("Content-Type", jsonType)
	requests, err := listFlowRequests(flowfunction)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(requests, "", "    ")
	w.Write(data)
	return
}

// requestTracesApiHandler request handler for traces of a request
func requestTracesApiHandler(w http.ResponseWriter, traceId string) {
	w.Header().Set("Content-Type", jsonType)
	trace, err := getRequestTraces(traceId)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(trace, "", "    ")
	w.Write(data)
	return
}

// listFlowFunctions request to list-flow-function to get flow-function list
func listFlowFunctions() ([]*Function, error) {
	var err error

	c := http.Client{}

	request, _ := http.NewRequest(http.MethodGet, gateway_url+"function/list-flow-functions", nil)
	response, err := c.Do(request)

	if err == nil {

		if response.Body != nil {
			defer response.Body.Close()
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				return nil, fmt.Errorf("failed to get function list, %v", bErr)
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

// listFlowRequests request to metrics function to get list of request for a flow function
func listFlowRequests(flow string) (map[string]string, error) {
	var err error

	c := http.Client{}
	url := gateway_url + "function/metrics?method=list&function=" + flow
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	response, err := c.Do(request)

	if err == nil {

		if response.Body != nil {
			defer response.Body.Close()
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				return nil, fmt.Errorf("failed to get request list, %v", bErr)
			}

			var requests map[string]string
			mErr := json.Unmarshal(bodyBytes, &requests)
			if mErr != nil {
				return nil, fmt.Errorf("failed to get request list, %v", mErr)
			}

			return requests, nil
		}
	}

	return nil, fmt.Errorf("failed to get requests list, %v", err)
}

// getRequestTraces request to metrics funcion to get list of traces for a request traceID
func getRequestTraces(requestTraceId string) (*RequestTrace, error) {
	var err error

	c := http.Client{}
	url := gateway_url + "function/metrics?method=traces&trace=" + requestTraceId
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	response, err := c.Do(request)
	if err == nil {
		if response.Body != nil {
			defer response.Body.Close()
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				return nil, fmt.Errorf("failed to get traces, %v", bErr)
			}

			trace := &RequestTrace{}
			mErr := json.Unmarshal(bodyBytes, trace)
			if mErr != nil {
				return nil, fmt.Errorf("failed to get traces, %v", mErr)
			}

			return trace, nil
		}
	}
	return nil, fmt.Errorf("failed to get traces, %v", err)
}

// getDag request to dot-generator for the dag dot graph
func getDag(function string) (string, error) {
	var err error

	c := http.Client{}

	request, _ := http.NewRequest(http.MethodGet, gateway_url+"function/dot-generator?function="+function, nil)
	response, err := c.Do(request)
	if err == nil {

		if response.Body != nil {
			defer response.Body.Close()
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

func lockFilePresent() bool {
	path := filepath.Join(os.TempDir(), ".lock")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func createLockFile() (string, error) {
	path := filepath.Join(os.TempDir(), ".lock")
	log.Printf("Writing lock-file to: %s\n", path)
	writeErr := ioutil.WriteFile(path, []byte{}, 0660)

	atomic.StoreInt32(&acceptingConnections, 1)

	return path, writeErr
}

func markUnhealthy() error {
	atomic.StoreInt32(&acceptingConnections, 0)

	path := filepath.Join(os.TempDir(), ".lock")
	log.Printf("Removing lock-file : %s\n", path)
	removeErr := os.Remove(path)
	return removeErr
}

// initialize globals
func initialize() error {
	public_uri = os.Getenv("gateway_public_uri")
	gateway_url = os.Getenv("gateway_url")
	gen = pagegen.Must(pagegen.ParseGlob("assets/templates/*.html"))
	return nil
}

// healthRequestHandler health check handler
func healthRequestHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if atomic.LoadInt32(&acceptingConnections) == 0 || lockFilePresent() == false {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))

			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// apiRequestHandler handles API reqiest
func apiRequestHandler(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("accept")

	// If API request
	if !strings.Contains(accept, "json") {
		http.Error(w, "failed to handle api request, must accept json", http.StatusBadRequest)
	}

	log.Printf("Serving api request")

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

	// handle request for listing faas-flow function
	case "list-flows":
		listFlowsApiHandler(w)
		return

	// handle request for details for a specific function
	case "flow-desc":
		flowDescApiHandler(w, msg.Function)
		return

	// handle request for list requests for a specific function
	case "list-requests":
		listFlowRequestsApiHandler(w, msg.Function)

	// handle request for listing trace for a specific request trace
	case "request-traces":
		requestTracesApiHandler(w, msg.TraceID)
		return

	}
	http.Error(w, "failed to handle request, method doesn't match", http.StatusBadRequest)
}

// dashboardPageHandler handle dashboard view
func dashboardPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request for dashboard view")

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	htmlObj := HtmlObject{PublicURL: public_uri, Functions: functions}

	err = gen.ExecuteTemplate(w, "dashboard", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}

}

// tracePageHandler handle tracing view
func tracePageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request for trace view")

	// Check the provided flow
	flows, ok := r.URL.Query()["flow"]
	if !ok || len(flows[0]) == 0 {
		http.Error(w, "failed to generate requested page, no flow specified", http.StatusBadRequest)
	}
	flow := flows[0]

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	requests, err := listFlowRequests(flow)
	if err != nil {
		log.Printf("failed to generate requested page, error: %v", err)
		requests = make(map[string]string)
	}

	htmlObj := HtmlObject{PublicURL: public_uri, Functions: functions, Flow: flow, Requests: requests}

	err = gen.ExecuteTemplate(w, "metrics", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}
}

// htmlPageHandler handler html page request
func htmlPageHandler(w http.ResponseWriter, r *http.Request) {
	flow := ""
	flows, ok := r.URL.Query()["flow"]
	if ok && len(flows[0]) > 0 {
		flow = flows[0]
	}
	if flow == "" {
		// Handle dashboard
		dashboardPageHandler(w, r)
	} else {
		// Handle trace
		tracePageHandler(w, r)
	}
}

// fileRequestHandler Static file request handler
func fileRequestHandler(w http.ResponseWriter, r *http.Request, file string) {
	filepath := "./assets/static/" + file
	log.Printf("Serving file %s", filepath)
	http.ServeFile(w, r, filepath)
}

// requestHandler handles dashboard view
func requestHandler(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("accept")

	// Check if file request
	files, ok := r.URL.Query()["file"]
	if ok && len(files[0]) > 0 {
		fileRequestHandler(w, r, files[0])
		return
	}

	// Check if api request
	if strings.Contains(accept, "json") {
		apiRequestHandler(w, r)
		return
	}

	// handle html request
	htmlPageHandler(w, r)
}

func main() {

	err := initialize()
	if err != nil {
		log.Fatal("failed to initialize the gateway, error: ", err.Error())
	}
	log.Printf("successfully initialized gateway")

	atomic.StoreInt32(&acceptingConnections, 0)

	// Every openfaas function request lands here
	http.HandleFunc("/", requestHandler)

	// Health Check api for kubernets and swarm
	http.HandleFunc("/_/health", healthRequestHandler())

	path, writeErr := createLockFile()
	if writeErr != nil {
		log.Panicf("Cannot write %s. Error: %s", path, writeErr.Error())
	}

	err = http.ListenAndServe(":8080", nil)
	markUnhealthy()
	log.Fatal(err)
}
