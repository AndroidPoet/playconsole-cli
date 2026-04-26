package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintMarkdownSliceOfStructs(t *testing.T) {
	var buf bytes.Buffer
	previousWriter := writer
	SetWriter(&buf)
	t.Cleanup(func() {
		SetWriter(previousWriter)
	})

	Setup("markdown", false, false)

	type row struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := Print([]row{{Name: "alpha", Value: "1"}, {Name: "beta", Value: "2"}}); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "| NAME | VALUE |") {
		t.Fatalf("expected markdown header, got %q", out)
	}
	if !strings.Contains(out, "| alpha | 1 |") {
		t.Fatalf("expected first markdown row, got %q", out)
	}
}

func TestSetupSupportsMarkdownAlias(t *testing.T) {
	Setup("md", false, false)
	if currentFormat != FormatMarkdown {
		t.Fatalf("expected markdown format, got %q", currentFormat)
	}
}
