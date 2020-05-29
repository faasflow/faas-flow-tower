package sdk

import (
	"fmt"
	"strings"
)

func FormatServiceName(owner, functionName string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(owner), functionName)
}

func CreateServiceURL(URL, suffix string) string {
	if len(suffix) == 0 {
		return URL
	}
	columns := strings.Count(URL, ":")
	//columns in URL with port are 2 i.e. http://url:port
	if columns == 2 {
		baseURL := URL[:strings.LastIndex(URL, ":")]
		port := URL[strings.LastIndex(URL, ":"):]
		return fmt.Sprintf("%s.%s%s", baseURL, suffix, port)
	}
	return fmt.Sprintf("%s.%s", URL, suffix)
}
