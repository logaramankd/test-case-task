package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func RunAllTestCases(code string, language string, testCases []TestCase) []TestResult {

	sandboxURL := os.Getenv("SANDBOX_SERVICE_URL")
	if sandboxURL == "" {
		log.Println("SANDBOX_SERVICE_URL not set. Using default http://localhost:3001")
		sandboxURL = "http://localhost:3001"
	} else {
		log.Println("Using SANDBOX_SERVICE_URL from env:", sandboxURL)
	}

	var results []TestResult
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, tc := range testCases {
		wg.Add(1)

		go func(tc TestCase) {
			defer wg.Done()

			payload := map[string]string{
				"code":     code,
				"input":    tc.Input,
				"language": language,
			}

			body, _ := json.Marshal(payload)

			resp, err := http.Post(sandboxURL+"/run", "application/json", bytes.NewBuffer(body))
			if err != nil {
				mu.Lock()
				results = append(results, TestResult{
					Input:    tc.Input,
					Expected: tc.Expected,
					Output:   "sandbox-service error: " + err.Error(),
					Passed:   false,
				})
				mu.Unlock()
				return
			}

			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			// Check generic error response
			var generic map[string]interface{}
			if err := json.Unmarshal(respBody, &generic); err == nil {
				if errMsg, ok := generic["error"].(string); ok && errMsg != "" {
					mu.Lock()
					results = append(results, TestResult{
						Input:    tc.Input,
						Expected: tc.Expected,
						Output:   errMsg,
						Passed:   false,
					})
					mu.Unlock()
					return
				}
			}

			var runResp struct {
				Stdout   string `json:"stdout"`
				Stderr   string `json:"stderr"`
				ExitCode int    `json:"exitCode"`
			}

			if err := json.Unmarshal(respBody, &runResp); err != nil {
				mu.Lock()
				results = append(results, TestResult{
					Input:    tc.Input,
					Expected: tc.Expected,
					Output:   "invalid sandbox-service response",
					Passed:   false,
				})
				mu.Unlock()
				return
			}

			if runResp.ExitCode != 0 {
				mu.Lock()
				results = append(results, TestResult{
					Input:    tc.Input,
					Expected: tc.Expected,
					Output:   strings.TrimSpace(runResp.Stderr),
					Passed:   false,
				})
				mu.Unlock()
				return
			}

			output := strings.TrimSpace(runResp.Stdout)
			passed := output == strings.TrimSpace(tc.Expected)

			mu.Lock()
			results = append(results, TestResult{
				Input:    tc.Input,
				Expected: tc.Expected,
				Output:   output,
				Passed:   passed,
			})
			mu.Unlock()

		}(tc)
	}

	// WAIT for all goroutines to finish
	wg.Wait()

	return results
}
