package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	totalRequests := 0
	for _, function := range functions {
		requests, err := listFlowRequests(function.Name)
		if err != nil {
			log.Printf("failed to get requests, error: %v", err)
			continue
		}
		totalRequests = totalRequests + len(requests)
	}

	dashboardSpec := &DashboardSpec{
		TotalFlows:     len(functions),
		ReadyFlows:     len(functions),
		TotalRequests:  totalRequests,
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

	requests, err := listFlowRequests(flowName)
	if err != nil {
		log.Printf("failed to get requests, error: %v", err)
		flowDesc.InvocationCount = 0
	} else {
		flowDesc.InvocationCount = float64(len(requests))
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

	requestsList := make(map[string]*RequestTrace)

	for request, traceId := range requests {
		requestsList[request], err = listRequestTraces(traceId)
		if err != nil {
			log.Printf("failed to get request traces for request %s, traceId %s, error: %v",
				request, traceId, err)
			requestsList[request] = &RequestTrace{
				TraceId: traceId,
			}
		}
		tracingEnabled = true
	}

	flowRequests := &FlowRequests{
		TracingEnabled: tracingEnabled,
		Flow:           flowName,
		Requests:       requestsList,
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
	currentRequestID := r.URL.Query().Get("request")

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

	requestsList := make(map[string]*RequestTrace)

	for request, traceId := range requests {
		requestsList[request], err = listRequestTraces(traceId)
		if err != nil {
			log.Printf("failed to get request traces for request %s, traceId %s, error: %v",
				request, traceId, err)
			requestsList[request] = &RequestTrace{
				TraceId: traceId,
			}
		}
		tracingEnabled = true
		if currentRequestID == "" {
			currentRequestID = request
		}
	}

	flowRequests := &FlowRequests{
		TracingEnabled:   tracingEnabled,
		Flow:             flowName,
		Requests:         requestsList,
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

		Traces: requestsList[currentRequestID],

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

	w.Header().Set("Content-Type", jsonType)
	functions, err := listFlowFunctions()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to handle list request, error: %v", err), http.StatusInternalServerError)
		return
	}
	data, _ := json.MarshalIndent(functions, "", "    ")
	w.Write(data)
}

// deleteFlowsHandler handle api request to delete flow function
func deleteFlowsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body == nil {
		http.Error(w, "invalid request, no content", 500)
		return
	}

	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	flowName := msg.FlowName
	log.Printf("deleting flow %s", flowName)

	err = deleteFlowFunction(flowName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
	return
}

// flowDesc request handler for a flow function
func flowDescHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body == nil {
		http.Error(w, "invalid request, no content", 500)
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

	if r.Body == nil {
		http.Error(w, "invalid request, no content", 500)
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

	if r.Body == nil {
		http.Error(w, "invalid request, no content", 500)
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
