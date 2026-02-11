package main

import (
	"encoding/json"
	"log"
	"net/http"
)


type SubmitRequest struct {
	QuesID   string `json:"quesId"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type TestResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Output   string `json:"output"`
	Passed   bool   `json:"passed"`
}

var questionStore = map[string][]TestCase{
	"1": { // Sum of Two Numbers
		{Input: "2 3", Expected: "5"},
		{Input: "10 20", Expected: "30"},
		{Input: "7 8", Expected: "15"},
		{Input: "100 200", Expected: "300"},
		{Input: "0 5", Expected: "5"},
	},

	"3": { // Multiply Two Numbers
		{Input: "2 3", Expected: "6"},
		{Input: "4 5", Expected: "20"},
		{Input: "6 7", Expected: "42"},
		{Input: "10 10", Expected: "100"},
		{Input: "9 0", Expected: "0"},
	},

	"7": { // Check Even or Odd (Print "Even" or "Odd")
		{Input: "2", Expected: "Even"},
		{Input: "3", Expected: "Odd"},
		{Input: "10", Expected: "Even"},
		{Input: "7", Expected: "Odd"},
		{Input: "0", Expected: "Even"},
	},

	"5": { // Factorial of Number
		{Input: "3", Expected: "6"},
		{Input: "4", Expected: "24"},
		{Input: "5", Expected: "120"},
		{Input: "1", Expected: "1"},
		{Input: "0", Expected: "1"},
	},
}



func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.QuesID == "" || req.Code == "" {
		http.Error(w, "quesId and code are required", http.StatusBadRequest)
		return
	}

	if req.Language != "go" {
		http.Error(w, "Only Go language is supported", http.StatusBadRequest)
		return
	}

	testCases, ok := questionStore[req.QuesID]
	if !ok {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	results := RunAllTestCases(req.Code, testCases)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	http.HandleFunc("/submit", submitHandler)

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
