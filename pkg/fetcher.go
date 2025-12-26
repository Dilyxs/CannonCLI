package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func Fetch(methodtype, givenurl, ContentType string, maxResponseTime time.Duration, body map[string]any) (*http.Response, time.Duration, error) {
	u, _ := url.Parse(givenurl)
	params := url.Values{}
	for k, v := range body {
		params.Add(k, fmt.Sprintf("%v", v))
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

func ReadJSONFileData(filepath string) (map[string]any, error) {
	body := make(map[string]any)
	if filepath != "" {
		details, err := os.ReadFile(filepath)
		if err != nil {
			return body, fmt.Errorf("file does not exist!: %v or no body file provided!", filepath)
		}
		if err := json.Unmarshal(details, &body); err != nil {
			return body, fmt.Errorf("file contains weird json, cannot decode!: %v", filepath)
		}
	}
	return body, nil
}

func ExtractResponse(r *http.Response) (map[string]any, error, bool) {
	response := make(map[string]any)
	isOk := true
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		isOk = false
		return response, err, isOk
	}
	return response, nil, isOk
}

type Response struct {
	Link     string
	Method   string
	Datajson map[string]any
	R        *http.Response
	Latency  time.Duration
	Response map[string]any
	IsOk     bool
}

func EvaluateFetching(method, link, filepathforJson string, TimeLimit time.Duration) (Response, error) {
	datajson, err := ReadJSONFileData(filepathforJson)
	if err != nil {
		return Response{}, err
	}
	r, latency, err := Fetch(method, link, "application/json", TimeLimit, datajson)
	if err != nil {
		return Response{}, err
	}
	response, err, isOk := ExtractResponse(r)
	if err != nil {
		return Response{}, err
	}
	FinalizedResponse := Response{
		link, method,
		datajson, r, latency, response, isOk,
	}
	return FinalizedResponse, nil
}
