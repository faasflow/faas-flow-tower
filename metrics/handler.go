package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type SpanItem struct {
	TraceID       string `json:"traceID"`
	SpanID        string `json:"spanID"`
	OperationName string `json:"operationName"`
	StartTime     int    `json:"startTime"`
	Duration      int    `json:"duration"`
	// Other can be added based on the needs
}

type TraceItem struct {
	TraceID string      `json:"traceID"`
	Spans   []*SpanItem `json:"spans"`
}

type Traces struct {
	Data []*TraceItem `json:"data"`
}

type SpanOps struct {
	TraceID       string `json:"traceID"`
	SpanID        string `json:"spanID"`
	OperationName string `json:"operationName"`
}

type RequestItem struct {
	TraceID string     `json:"traceID"`
	Spans   []*SpanOps `json:"spans"`
}

type Requests struct {
	Data []*RequestItem `json:"data"`
}

type NodeTrace struct {
	StartTime int `json:"startTime"`
	Duration  int `json:"duration"`
}

type RequestTrace struct {
	RequestID  string
	NodeTraces map[string]*NodeTrace
	StartTime  int
	Duration   int
}

var (
	trace_url = ""
)

func listRequest(function string) (string, error) {
	resp, err := http.Get(trace_url + "api/traces?service=" + function)
	if err != nil {
		return "", fmt.Errorf("failed to request trace service, error %v ", err)
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		return "", fmt.Errorf("failed to request trace service, status code %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read trace result, read error %v", err)
	}

	if len(bodyBytes) == 0 {
		return "", fmt.Errorf("failed to get request traces, empty result")
	}

	requests := &Requests{}
	err = json.Unmarshal(bodyBytes, requests)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal requests lists, error %v", err)
	}

	requestMap := make(map[string]string)
	for _, request := range requests.Data {
		if request.Spans == nil {
			continue
		}
		for _, span := range request.Spans {
			if span.TraceID == request.TraceID && span.TraceID == span.SpanID {
				requestMap[span.OperationName] = request.TraceID
				break
			}
		}
	}

	encoded, err := json.MarshalIndent(requestMap, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to encode request list, error %v", err)
	}

	return string(encoded), nil
}

func listTraces(request string) (string, error) {
	resp, err := http.Get(trace_url + "api/traces/" + request)
	if err != nil {
		return "", fmt.Errorf("failed to request trace service, error %v ", err)
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		return "", fmt.Errorf("failed to request trace service, status code %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read trace result, read error %v", err)
	}

	if len(bodyBytes) == 0 {
		return "", fmt.Errorf("failed to get request traces, empty result")
	}

	traces := &Traces{}
	err = json.Unmarshal(bodyBytes, traces)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal requests lists, error %v", err)
	}

	if traces.Data == nil || len(traces.Data) == 0 {
		return "", fmt.Errorf("failed to get request traces, empty data")
	}

	requestTrace := traces.Data[0]
	if requestTrace.TraceID != request {
		return "", fmt.Errorf("invalid request trace %s", requestTrace.TraceID)
	}

	response := &RequestTrace{}
	response.RequestID = request
	response.NodeTraces = make(map[string]*NodeTrace)

	var lastSpanStart int
	var lastSpanDuration int

	for _, span := range requestTrace.Spans {
		if span.TraceID == request && span.TraceID == span.SpanID {
			response.StartTime = span.StartTime
			lastSpanStart = span.StartTime
		} else {
			node := &NodeTrace{}
			node.StartTime = span.StartTime
			node.Duration = span.Duration
			response.NodeTraces[span.OperationName] = node
			if span.StartTime > lastSpanStart {
				lastSpanStart = span.StartTime
				lastSpanDuration = span.Duration
			}
		}
	}
	response.Duration = lastSpanStart - response.StartTime + lastSpanDuration

	encoded, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to encode request list, error %v", err)
	}

	return string(encoded), nil
}

// Handle a serverless request
func Handle(req []byte) string {
	values, err := url.ParseQuery(os.Getenv("Http_Query"))
	if err != nil {
		log.Fatal("No argument specified")
	}

	method := values.Get("method")
	if method == "" {
		method = "list"
	}

	trace_url = os.Getenv("trace_url")
	if trace_url == "" {
		trace_url = "http://jaegertracing:16686/"
	}

	var resp string

	switch method {

	case "list":
		function := values.Get("function")
		if function == "" {
			log.Fatal("No function specified")
		}
		resp, err = listRequest(function)

	case "traces":
		trace := values.Get("trace")
		if len(trace) <= 0 {
			log.Fatal("No request specified")
		}
		resp, err = listTraces(trace)
	}

	if err != nil {
		log.Fatal("Failed to process, error ", err)
	}

	return resp
}
