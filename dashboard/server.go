package main

import (
	"fmt"
	pageGen "html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	jsonType = "application/json"
	htmlType = "text/html"
)

var (
	publicUri                    = ""
	gen        *pageGen.Template = nil
	gatewayUrl                   = ""
)

// initialize globals
func initialize() error {
	publicUri = os.Getenv("gateway_public_uri")
	if len(publicUri) == 0 {
		publicUri = "/function/faas-flow-dashboard"
	}
	gatewayUrl = os.Getenv("gateway_url")
	gen = pageGen.Must(pageGen.ParseGlob("assets/templates/*.html"))
	return nil
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}

func main() {

	readTimeout := parseIntOrDurationValue(os.Getenv("read_timeout"), 10*time.Second)
	writeTimeout := parseIntOrDurationValue(os.Getenv("write_timeout"), 10*time.Second)

	var err error

	err = initialize()
	if err != nil {
		log.Fatal("failed to initialize the gateway, error: ", err.Error())
	}
	log.Printf("successfully initialized gateway")

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	// Template
	http.HandleFunc("/", dashboardPageHandler)
	http.HandleFunc("/flow/info", flowInfoPageHandler)
	http.HandleFunc("/flow/requests", flowRequestsPageHandler)
	http.HandleFunc("/flow/request/monitor", flowRequestMonitorPageHandler)

	// Static content
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./assets/static/"))))

	// API request
	http.HandleFunc("/api/flow/list", listFlowsHandler)
	http.HandleFunc("/api/flow/info", flowDescHandler)
	http.HandleFunc("/api/flow/requests", listFlowRequestsHandler)
	http.HandleFunc("/api/flow/request/traces", listFlowRequestsHandler)

	log.Fatal(s.ListenAndServe())
}
