package function

import (
	"log"
	"net/url"
	"os"
)

var (
	trace_url = ""
)

func listRequest(function string) (string, error) {
	return "", nil
}

func listTraces(function string, request string) (string, error) {
	return "", nil
}

// Handle a serverless request
func Handle(req []byte) string {
	values, err := url.ParseQuery(os.Getenv("Http_Query"))
	if err != nil {
		log.Fatal("No argument specified")
	}

	function := values.Get("function")
	if function == "" {
		log.Fatal("No function specified")
	}

	method := values.Get("method")
	if method == "" {
		method = "list-requests"
	}

	trace_url = os.Getenv("trace_url")
	if trace_url == "" {
		trace_url = "http://localhost:16686/"
	}

	var resp string

	switch method {

	case "list-requests":
		resp, err = listRequest(function)

	case "traces":
		request := values.Get("request")
		if len(request) <= 0 {
			log.Fatal("No request specified")
		}
		resp, err = listTraces(function, request)
	}

	if err != nil {
		log.Fatal("Failed to process, error ", err)
	}

	return resp
}
