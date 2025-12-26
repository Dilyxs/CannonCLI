package pkg

import (
	"net/http"
	"net/url"
	"time"
)

func Fetch(methodtype, givenurl, ContentType string, maxResponseTime time.Duration, body map[string]any) (*http.Response, time.Duration, error) {
	u, _ := url.Parse(givenurl)
	params := url.Values{}
	for k, v := range body {
		params.Add(k, v.(string))
	}
	u.RawQuery = params.Encode()
	req, err := http.NewRequest(methodtype, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-type", ContentType)
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: maxResponseTime,
	}
	currenttime := time.Now()
	response, err := client.Do(req)
	responsetime := time.Now()
	if err != nil {
		return nil, 0, err
	}
	latency := responsetime.Sub(currenttime)
	return response, latency, nil
}
