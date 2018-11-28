package function

import (
	"encoding/json"
	"github.com/openfaas/openfaas-cloud/sdk"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Handle a serverless request
func Handle(req []byte) string {

	gatewayURL := os.Getenv("gateway_url")

	c := http.Client{
		Timeout: time.Second * 3,
	}

	httpReq, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/functions", nil)

	basicAuth := os.Getenv("basic_auth")
	if strings.ToUpper(basicAuth) == "TRUE" {
		addAuthErr := sdk.AddBasicAuth(httpReq)
		if addAuthErr != nil {
			log.Printf("Basic auth error %s", addAuthErr)
		}
	}

	response, err := c.Do(httpReq)
	if err != nil {
		log.Fatal(err)
	}

	filtered := []function{}

	defer response.Body.Close()
	bodyBytes, bErr := ioutil.ReadAll(response.Body)
	if bErr != nil {
		log.Fatal(bErr)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("unable to query functions, status: %d, message: %s", response.StatusCode, string(bodyBytes))
	}

	functions := []function{}
	mErr := json.Unmarshal(bodyBytes, &functions)
	if mErr != nil {
		log.Fatal(mErr)
	}

	for _, fn := range functions {
		for k, v := range fn.Labels {
			if k == "faas-flow" && v == "1" {
				filtered = append(filtered, fn)
			}
		}
	}

	bytesOut, _ := json.Marshal(filtered)
	return string(bytesOut)
}

type function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}
