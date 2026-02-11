package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type TestCase struct {
	Input    string `bson:"input"`
	Expected string `bson:"expected"`
}

func RunAllTestCases(code string, testCases []TestCase) []TestResult {
	sourceFile := "user_solution.go"
	binaryFile := "user_solution_bin"

	if runtime.GOOS == "windows" {
		binaryFile += ".exe"
	}

	os.WriteFile(sourceFile, []byte(code), 0644)
	defer os.Remove(sourceFile)
	defer os.Remove(binaryFile)

	// Compile once
	buildCmd := exec.Command("go", "build", "-o", binaryFile, sourceFile)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		// Compilation failed
		return []TestResult{
			{
				Input:    "",
				Expected: "",
				Output:   string(buildOutput),
				Passed:   false,
			},
		}
	}

	var wg sync.WaitGroup

	resultsChan := make(chan TestResult, len(testCases))

	for _, tc := range testCases {
		wg.Add(1)

		go func(tc TestCase) {
			defer wg.Done()
			result := runSingleTest(binaryFile, tc)
			resultsChan <- result
		}(tc)
	}

	wg.Wait()
	close(resultsChan)

	var results []TestResult
	for r := range resultsChan {
		results = append(results, r)
	}

	return results
}

func runSingleTest(filename string, tc TestCase) TestResult {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, ".\\"+filename)
	} else {
		cmd = exec.CommandContext(ctx, "./"+filename)
	}

	cmd.Stdin = strings.NewReader(tc.Input)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	output := strings.TrimSpace(out.String())

	if err != nil {
		output = err.Error() + " | " + output
	}

	passed := false
	if err == nil && output == tc.Expected {
		passed = true
	}

	return TestResult{
		Input:    tc.Input,
		Expected: tc.Expected,
		Output:   output,
		Passed:   passed,
	}
}
