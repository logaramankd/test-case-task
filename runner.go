package main

import (
	"bytes"
	"encoding/json"
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

    body, _ := json.Marshal(payload)

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

    var results []TestResult
    json.NewDecoder(resp.Body).Decode(&results)

    return results
}

