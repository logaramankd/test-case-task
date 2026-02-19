package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func SolveWithAgent(quesID string, language string) map[string]interface{} {

	testCases := questionStore[quesID]
	if len(testCases) == 0 {
		return map[string]interface{}{
			"error": "No test cases found",
		}
	}

	maxAttempts := 5
	var lastCode string
	var lastResults []TestResult

	for attempt := 1; attempt <= maxAttempts; attempt++ {

		log.Println("==========")
		log.Println("Attempt:", attempt)

		// 1️⃣ Generate or fix code
		code := GenerateCodeWithOllama(quesID, language, lastResults)
		lastCode = code

		log.Println("Generated Code:\n", code)

		// 2️⃣ Run tests
		results := RunAllTestCases(code, language, testCases)
		lastResults = results

		log.Println("Test Results:", results)

		// 3️⃣ Validate results
		allPassed := true

		if len(results) == 0 {
			allPassed = false
		}

		for _, r := range results {
			if !r.Passed {
				allPassed = false
				break
			}
		}

		if allPassed {
			log.Println("All tests passed.")
			return map[string]interface{}{
				"attempts":  attempt,
				"finalCode": code,
				"results":   results,
				"success":   true,
			}
		}
	}

	log.Println("Max attempts reached.")

	return map[string]interface{}{
		"attempts":  maxAttempts,
		"finalCode": lastCode,
		"results":   lastResults,
		"success":   false,
	}
}

func GenerateCodeWithOllama(quesID, language string, previousResults []TestResult) string {

	problemPrompt := BuildPrompt(quesID, language, previousResults)

	payload := map[string]interface{}{
		"model":  "llama3.2",
		"prompt": problemPrompt,
		"stream": false,
	}

	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Println("Ollama error:", err)
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Decode error:", err)
		return ""
	}

	return cleanCode(result.Response)
}

func BuildPrompt(quesID, language string, previousResults []TestResult) string {

	baseProblem := ""

	switch quesID {
	case "1":
		baseProblem = "Read two integers from standard input separated by space and print their sum."
	case "3":
		baseProblem = "Read two integers from standard input separated by space and print their product."
	case "7":
		baseProblem = "Read an integer and print Even if number is even else print Odd."
	case "5":
		baseProblem = "Read an integer and print its factorial."
	}

	prompt := ""
	prompt += "You are a competitive programming expert.\n"
	prompt += "Return ONLY raw " + language + " code.\n"
	prompt += "Do NOT include markdown.\n"
	prompt += "Do NOT include explanation.\n"
	prompt += "Read input from standard input.\n\n"
	prompt += "Problem:\n" + baseProblem + "\n\n"

	if len(previousResults) > 0 {
		prompt += "Previous attempt failed. Fix the code.\n"

		for _, r := range previousResults {
			if !r.Passed {
				prompt += "Input: " + r.Input + "\n"
				prompt += "Expected: " + r.Expected + "\n"
				prompt += "Got: " + r.Output + "\n\n"
			}
		}
	}

	return prompt
}

func cleanCode(code string) string {
	code = strings.TrimSpace(code)

	// Remove markdown blocks
	if strings.Contains(code, "```") {
		parts := strings.Split(code, "```")
		for _, p := range parts {
			p = strings.TrimSpace(p)

			if strings.HasPrefix(p, "go") {
				return strings.TrimSpace(strings.TrimPrefix(p, "go"))
			}
			if strings.HasPrefix(p, "python") {
				return strings.TrimSpace(strings.TrimPrefix(p, "python"))
			}
			if strings.HasPrefix(p, "java") {
				return strings.TrimSpace(strings.TrimPrefix(p, "java"))
			}
			if strings.HasPrefix(p, "javascript") {
				return strings.TrimSpace(strings.TrimPrefix(p, "javascript"))
			}
		}
	}

	return code
}
