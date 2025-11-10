package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func parseKeyValuePairs(s string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		result[key] = val
	}

	return result
}

func parseHeaderStrings(s string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		result[key] = val
	}

	return result
}

func executeHttpRequest(reqData *RequestData) (string, error) {
	client := &http.Client{}

	url := reqData.URL
	params := parseKeyValuePairs(reqData.Params)

	if len(params) > 0 {
		var q []string
		for k, v := range params {
			q = append(q, fmt.Sprintf("%s=%s", k, v))
		}
		if strings.Contains(url, "?") {
			url += "&" + strings.Join(q, "&")
		} else {
			url += "?" + strings.Join(q, "&")
		}
	}

	var body io.Reader
	if reqData.Body != "" {
		body = bytes.NewBufferString(reqData.Body)
	}

	req, err := http.NewRequest(reqData.Method, url, body)
	if err != nil {
		return "", err
	}

	headers := parseHeaderStrings(reqData.Headers)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if strings.HasPrefix(strings.ToLower(reqData.Auth), "bearer ") {
		req.Header.Set("Authorization", reqData.Auth)
	} else if strings.Contains(reqData.Auth, ":") {
		parts := strings.SplitN(reqData.Auth, ":", 2)
		req.SetBasicAuth(parts[0], parts[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	if json.Unmarshal(respBytes, &map[string]interface{}{}) == nil {
		json.Indent(&prettyJSON, respBytes, "", " ")
		return prettyJSON.String(), nil
	}

	return string(respBytes), nil
}
