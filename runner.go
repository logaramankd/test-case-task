package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func RunAllTestCases(code string, testCases []TestCase) []TestResult {

	tempDir, err := os.MkdirTemp("", "submission-*")
	if err != nil {
		return []TestResult{
			{Output: "Failed to create temp directory", Passed: false},
		}
	}
	defer os.RemoveAll(tempDir)

	sourceFile := filepath.Join(tempDir, "solution.go")
	binaryFile := filepath.Join(tempDir, "solution_bin")

	if runtime.GOOS == "windows" {
		binaryFile += ".exe"
	}

	if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return []TestResult{
			{Output: "Failed to write source file", Passed: false},
		}
	}

	buildCmd := exec.Command("go", "build", "-o", binaryFile, sourceFile)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		return []TestResult{
			{
				Output: string(buildOutput),
				Passed: false,
			},
		}
	}

	results := make([]TestResult, 0, len(testCases))

	for _, tc := range testCases {
		result := runSingleTest(binaryFile, tc)
		results = append(results, result)
	}

	return results
}

func runSingleTest(binaryPath string, tc TestCase) TestResult {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, binaryPath)
	} else {
		cmd = exec.CommandContext(ctx, binaryPath)
	}

	cmd.Stdin = strings.NewReader(tc.Input)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	output := strings.TrimSpace(out.String())

	if ctx.Err() == context.DeadlineExceeded {
		return TestResult{
			Input:    tc.Input,
			Expected: tc.Expected,
			Output:   "Time Limit Exceeded",
			Passed:   false,
		}
	}

	if err != nil {
		output = err.Error() + " | " + output
	}

	passed := err == nil && output == strings.TrimSpace(tc.Expected)

	return TestResult{
		Input:    tc.Input,
		Expected: tc.Expected,
		Output:   output,
		Passed:   passed,
	}
}
