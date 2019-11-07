package main

// Function object to retrieve and response flow-function details
type Function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}

type DashboardSpec struct {
	TotalFlows     int
	ReadyFlows     int
	TotalRequests  int
	ActiveRequests int
}

type Location struct {
	Name string
	Link string
}

type FlowDesc struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	Description     string            `json:"description"`
	InvocationCount float64           `json:"invocation-count"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Dot             string            `json:"dot,omitempty"`
}

type FlowRequests struct {
	Flow             string
	TracingEnabled   bool
	Requests         map[string]*RequestTrace
	CurrentRequestID string
}

// NodeTrace traces of each nodes in a dag
type NodeTrace struct {
	StartTime int `json:"start-time"`
	Duration  int `json:"duration"`
	// Other can be added based on the needs
}

// RequestTrace object to retrieve and response traces details
type RequestTrace struct {
	RequestID  string                `json:"request-id"`
	TraceId    string                `json:"trace-id"`
	NodeTraces map[string]*NodeTrace `json:"traces"`
	StartTime  int                   `json:"start-time"`
	Duration   int                   `json:"duration"`
	Status     string                `json:"status"`
}
