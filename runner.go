package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func RunAllTestCases(code string, language string, testCases []TestCase) []TestResult {

	sandboxURL := os.Getenv("SANDBOX_SERVICE_URL")
	if sandboxURL == "" {
		sandboxURL = "http://localhost:3001"
	}

	payload := map[string]interface{}{
		"code":      code,
		"language":  language,
		"testCases": testCases,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return []TestResult{
			{Output: "internal error: failed to encode request", Passed: false},
		}
	}

	resp, err := http.Post(sandboxURL+"/run", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return []TestResult{
			{
				Output: "sandbox-service error: " + err.Error(),
				Passed: false,
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := string(bodyBytes)
		if errMsg == "" {
			errMsg = "sandbox-service returned status " + resp.Status
		}
		return []TestResult{
			{
				Output: "sandbox-service error: " + errMsg,
				Passed: false,
			},
		}
	}

	var results []TestResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return []TestResult{
			{
				Output: "sandbox-service error: invalid response: " + err.Error(),
				Passed: false,
			},
		}
	}

	return results
}
