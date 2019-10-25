package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HtmlObject object to render web page
type HtmlObject struct {
	PublicURL string
	Functions []*Function

	LocationDepths []*Location

	CurrentLocation *Location

	InnerHtml string

	DashBoard *DashboardSpec
	Flow      *FlowDesc
	Requests  *FlowRequests
	Traces    *RequestTrace
}

// Message API request query
type Message struct {
	FlowName string `json:"function"`
	TraceID  string `json:"trace-id"`
}

// dashboardPageHandler handle dashboard view
func dashboardPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request for dashboard view")

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	dashboardSpec := &DashboardSpec{
		TotalFlows:     len(functions),
		ReadyFlows:     len(functions),
		TotalRequests:  0,
		ActiveRequests: 0,
	}

	htmlObj := HtmlObject{
		PublicURL: publicUri,
		Functions: functions,

		InnerHtml: "dashboard",

		DashBoard: dashboardSpec,
	}

	err = gen.ExecuteTemplate(w, "index", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}

}

// flowInfoPageHandler handle flow info page view
func flowInfoPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request for dashboard view")

	flowName := r.URL.Query().Get("flow-name")

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	flowDesc, err := buildFlowDesc(functions, flowName)
	if err != nil {
		log.Printf("failed to get function desc, error: %v", err)
	}

	htmlObj := HtmlObject{
		PublicURL: publicUri,
		Functions: functions,

		CurrentLocation: &Location{
			Name: "Flow : " + flowName + "",
			Link: "/function/faas-flow-dashboard/flow/info?flow-name=" + flowName,
		},

		InnerHtml: "flow-info",

		Flow: flowDesc,
	}

	err = gen.ExecuteTemplate(w, "index", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}
}

// flowRequestsPageHandler handle tracing view
func flowRequestsPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request for request list view")

	flowName := r.URL.Query().Get("flow-name")

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	tracingEnabled := false
	requests, err := listFlowRequests(flowName)
	if err != nil {
		log.Printf("failed to get requests, error: %v", err)
		requests = make(map[string]string)
	}

	for range requests {
		tracingEnabled = true
		break
	}

	flowRequests := &FlowRequests{
		TracingEnabled: tracingEnabled,
		Flow:           flowName,
		Requests:       requests,
	}

	locationDepths := []*Location{
		&Location{
			Name: "Flow : " + flowName + "",
			Link: "/function/faas-flow-dashboard/flow/info?flow-name=" + flowName,
		},
	}

	htmlObj := HtmlObject{
		PublicURL: publicUri,
		Functions: functions,

		LocationDepths: locationDepths,

		CurrentLocation: &Location{
			Name: "Requests",
			Link: "/function/faas-flow-dashboard/flow/requests?flow-name=" + flowName,
		},

		Requests: flowRequests,

		InnerHtml: "requests",
	}

	err = gen.ExecuteTemplate(w, "index", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}
}

// flowRequestMonitorPageHandler handle tracing view
func flowRequestMonitorPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request for request monitor view")

	flowName := r.URL.Query().Get("flow-name")

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	tracingEnabled := false
	requests, err := listFlowRequests(flowName)
	if err != nil {
		log.Printf("failed to get requests, error: %v", err)
		requests = make(map[string]string)
	}

	currentRequestID := ""

	for currentRequestID, _ = range requests {
		tracingEnabled = true
		break
	}

	flowRequests := &FlowRequests{
		TracingEnabled:   tracingEnabled,
		Flow:             flowName,
		Requests:         requests,
		CurrentRequestID: currentRequestID,
	}

	locationDepths := []*Location{
		&Location{
			Name: "Flow : " + flowName + "",
			Link: "/function/faas-flow-dashboard/flow/info?flow-name=" + flowName,
		},
		&Location{
			Name: "Requests",
			Link: "/function/faas-flow-dashboard/flow/requests?flow-name=" + flowName,
		},
	}

	htmlObj := HtmlObject{
		PublicURL: publicUri,
		Functions: functions,

		LocationDepths: locationDepths,

		CurrentLocation: &Location{
			Name: "requests-choice",
		},

		Requests: flowRequests,

		Traces: &RequestTrace{
			RequestID: currentRequestID,

			// TODO: initialize
			Duration: 0,
			Status:   "unknown",
		},

		InnerHtml: "request-monitor",
	}

	err = gen.ExecuteTemplate(w, "index", htmlObj)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate requested page, error: %v", err), http.StatusInternalServerError)
	}
}

// API

// listFlows handle api request to list flow function
func listFlowsHandler(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("accept")

	// If API request
	if !strings.Contains(accept, "json") {
		http.Error(w, "failed to handle api request, must accept json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", jsonType)
	functions, err := listFlowFunctions()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle list request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(functions, "", "    ")
	w.Write(data)
}

// flowDesc request handler for a flow function
func flowDescHandler(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("accept")

	// If API request
	if !strings.Contains(accept, "json") {
		http.Error(w, "failed to handle api request, must accept json", http.StatusBadRequest)
	}

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

	flowName := msg.FlowName

	functions, err := listFlowFunctions()
	if err != nil {
		log.Printf("failed to get functions, error: %v", err)
		functions = make([]*Function, 0)
	}

	flowDesc, err := buildFlowDesc(functions, flowName)
	if err != nil {
		http.Error(w, "failed to handle request, "+err.Error(), 500)
		return
	}

	data, _ := json.MarshalIndent(flowDesc, "", "    ")
	w.Header().Set("Content-Type", jsonType)
	w.Write(data)
}

// listFlowRequestsApiHandler list the requests for a flow function
func listFlowRequestsHandler(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("accept")

	// If API request
	if !strings.Contains(accept, "json") {
		http.Error(w, "failed to handle api request, must accept json", http.StatusBadRequest)
	}

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

	flowFunction := msg.FlowName

	w.Header().Set("Content-Type", jsonType)
	requests, err := listFlowRequests(flowFunction)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(requests, "", "    ")
	w.Write(data)
	return
}

// requestTracesApiHandler request handler for traces of a request
func requestTracesHandler(w http.ResponseWriter, r *http.Request) {

	accept := r.Header.Get("accept")

	// If API request
	if !strings.Contains(accept, "json") {
		http.Error(w, "failed to handle api request, must accept json", http.StatusBadRequest)
	}

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

	traceID := msg.TraceID

	w.Header().Set("Content-Type", jsonType)
	trace, err := listRequestTraces(traceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(trace, "", "    ")
	w.Write(data)
	return
}
