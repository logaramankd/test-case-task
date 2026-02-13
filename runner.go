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

func RunAllTestCases(code string, language string, testCases []TestCase) []TestResult {

	tempDir, err := os.MkdirTemp("", "submission-*")
	if err != nil {
		return []TestResult{
			{Output: "Failed to create temp directory", Passed: false},
		}
	}
	defer os.RemoveAll(tempDir)

	var executablePath string

	switch language {

	case "go":
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
				{Output: string(buildOutput), Passed: false},
			}
		}

		executablePath = binaryFile

	case "python":
		sourceFile := filepath.Join(tempDir, "solution.py")

		if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
			return []TestResult{
				{Output: "Failed to write source file", Passed: false},
			}
		}

		executablePath = sourceFile

	default:
		return []TestResult{
			{Output: "Unsupported language", Passed: false},
		}
	}

	results := make([]TestResult, 0, len(testCases))

	for _, tc := range testCases {
		result := runSingleTest(language, executablePath, tc)
		results = append(results, result)
	}

	return results
}

func runSingleTest(language string, executablePath string, tc TestCase) TestResult {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	switch language {
	case "go":
		cmd = exec.CommandContext(ctx, executablePath)
	case "python":
		// Use python3 if your system requires it
		cmd = exec.CommandContext(ctx, "python", executablePath)
	default:
		return TestResult{
			Input:    tc.Input,
			Expected: tc.Expected,
			Output:   "Unsupported language",
			Passed:   false,
		}
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
