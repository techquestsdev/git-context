package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
)

var captureMutex sync.Mutex

func captureOutput(f func()) string {
	// Lock to prevent parallel tests from interfering with each other
	captureMutex.Lock()
	defer captureMutex.Unlock()

	// Capture both color.Output and os.Stdout
	oldColor := color.Output
	oldStdout := os.Stdout
	oldNoColor := color.NoColor

	r, w, _ := os.Pipe()
	color.Output = w
	os.Stdout = w

	// Disable color for testing to get clean output
	color.NoColor = true

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		_, _ = io.Copy(&buf, r)

		outC <- buf.String()
	}()

	f()

	_ = w.Close()

	color.Output = oldColor
	os.Stdout = oldStdout
	color.NoColor = oldNoColor
	out := <-outC

	return out
}

func TestPrintSuccess(t *testing.T) {
	t.Parallel()

	output := captureOutput(func() {
		PrintSuccess("Test success message")
	})

	if !strings.Contains(output, "✓") {
		t.Error("Success message should contain checkmark symbol")
	}

	if !strings.Contains(output, "Test success message") {
		t.Error("Success message should contain the message text")
	}
}

func TestPrintError(t *testing.T) {
	t.Parallel()

	output := captureOutput(func() {
		PrintError("Test error message")
	})

	if !strings.Contains(output, "✗") {
		t.Error("Error message should contain X symbol")
	}

	if !strings.Contains(output, "Test error message") {
		t.Error("Error message should contain the message text")
	}
}

func TestPrintWarning(t *testing.T) {
	t.Parallel()

	output := captureOutput(func() {
		PrintWarning("Test warning message")
	})

	if !strings.Contains(output, "⚠") {
		t.Error("Warning message should contain warning symbol")
	}

	if !strings.Contains(output, "Test warning message") {
		t.Error("Warning message should contain the message text")
	}
}

func TestPrintInfo(t *testing.T) {
	t.Parallel()

	output := captureOutput(func() {
		PrintInfo("Test info message")
	})

	if !strings.Contains(output, "ℹ") {
		t.Error("Info message should contain info symbol")
	}

	if !strings.Contains(output, "Test info message") {
		t.Error("Info message should contain the message text")
	}
}

func TestPrintHeader(t *testing.T) {
	t.Parallel()

	output := captureOutput(func() {
		PrintHeader("Test Header")
	})

	if !strings.Contains(output, "===") {
		t.Error("Header should contain === markers")
	}

	if !strings.Contains(output, "Test Header") {
		t.Error("Header should contain the title text")
	}
}

func TestPrint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		outputType OutputType
		message    string
	}{
		{"Success", OutputSuccess, "Success message"},
		{"Error", OutputError, "Error message"},
		{"Warning", OutputWarning, "Warning message"},
		{"Info", OutputInfo, "Info message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := captureOutput(func() {
				Print(tt.outputType, tt.message)
			})

			if !strings.Contains(output, tt.message) {
				t.Errorf("Output should contain message: %s", tt.message)
			}
		})
	}
}

func TestPrintf(t *testing.T) {
	t.Parallel()

	output := captureOutput(func() {
		Printf(OutputInfo, "User: %s, Email: %s", "Test User", "test@example.com")
	})

	if !strings.Contains(output, "User: Test User") {
		t.Error("Printf should format the message correctly")
	}

	if !strings.Contains(output, "Email: test@example.com") {
		t.Error("Printf should include all formatted values")
	}
}

func TestPrintTable(t *testing.T) {
	t.Parallel()

	headers := []string{"Name", "Email", "Status"}
	rows := [][]string{
		{"work", "work@example.com", "● (active)"},
		{"personal", "personal@example.com", ""},
	}

	output := captureOutput(func() {
		PrintTable(headers, rows)
	})

	// Verify headers
	if !strings.Contains(output, "Name") {
		t.Error("Table should contain Name header")
	}

	if !strings.Contains(output, "Email") {
		t.Error("Table should contain Email header")
	}

	if !strings.Contains(output, "Status") {
		t.Error("Table should contain Status header")
	}

	// Verify rows
	if !strings.Contains(output, "work") {
		t.Error("Table should contain work profile")
	}

	if !strings.Contains(output, "work@example.com") {
		t.Error("Table should contain work email")
	}

	if !strings.Contains(output, "personal") {
		t.Error("Table should contain personal profile")
	}

	if !strings.Contains(output, "personal@example.com") {
		t.Error("Table should contain personal email")
	}

	// Verify separator line (dashes)
	if !strings.Contains(output, "---") {
		t.Error("Table should contain separator line")
	}
}

func TestPrintTableEmptyRows(t *testing.T) {
	t.Parallel()

	headers := []string{"Name", "Email"}
	rows := [][]string{}

	output := captureOutput(func() {
		PrintTable(headers, rows)
	})

	// Should still print headers
	if !strings.Contains(output, "Name") {
		t.Error("Table should contain headers even with empty rows")
	}
}

func TestPrintTableColumnWidthCalculation(t *testing.T) {
	t.Parallel()

	headers := []string{"Short", "VeryLongHeaderName"}
	rows := [][]string{
		{"A", "B"},
		{"VeryLongValue", "ShortVal"},
	}

	output := captureOutput(func() {
		PrintTable(headers, rows)
	})

	// Verify content is present
	if !strings.Contains(output, "VeryLongHeaderName") {
		t.Error("Table should contain long header")
	}

	if !strings.Contains(output, "VeryLongValue") {
		t.Error("Table should contain long value")
	}
}
