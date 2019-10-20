package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// listFlowFunctions request to list-flow-function to get flow-function list
func listFlowFunctions() ([]*Function, error) {
	var err error

	c := http.Client{}

	request, _ := http.NewRequest(http.MethodGet, gatewayUrl+"function/list-flow-functions", nil)
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

// getDot request to dot-generator for the dag dot graph
func getDot(function string) (string, error) {
	var err error

	c := http.Client{}

	request, _ := http.NewRequest(http.MethodGet, gatewayUrl+"function/dot-generator?function="+function, nil)
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

// listFlowRequests request to metrics function to get list of request for a flow function
func listFlowRequests(flow string) (map[string]string, error) {
	var err error

	c := http.Client{}
	url := gatewayUrl + "function/metrics?method=list&function=" + flow
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

// buildFlowDesc get a flow details
func buildFlowDesc(functions []*Function, flowName string) (*FlowDesc, error) {

	var functionObj *Function
	for _, functionObj = range functions {
		if functionObj.Name == flowName {
			break
		}
	}

	description := functionObj.Annotations["faas-flow-desc"]

	dot, dErr := getDot(flowName)
	if dErr != nil {
		return nil, fmt.Errorf("failed to get dot, %v", dErr)
	}

	flowDesc := &FlowDesc{
		Name:            functionObj.Name,
		Image:           functionObj.Image,
		Description:     description,
		InvocationCount: functionObj.InvocationCount,
		Replicas:        functionObj.Replicas,
		Labels:          functionObj.Labels,
		Annotations:     functionObj.Annotations,
		Dot:             dot,
	}

	return flowDesc, nil
}

// listRequestTraces request to metrics function to get list of traces for a request traceID
func listRequestTraces(requestTraceId string) (*RequestTrace, error) {
	var err error

	c := http.Client{}
	url := gatewayUrl + "function/metrics?method=traces&trace=" + requestTraceId
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
