package output

import (
	"bytes"
	"strings"
	"testing"
)

// Status messages must never land on stdout, otherwise `gpc ... | jq` breaks.
func TestStatusMessagesDoNotWriteToDataWriter(t *testing.T) {
	var buf bytes.Buffer
	previousWriter := writer
	SetWriter(&buf)
	t.Cleanup(func() { SetWriter(previousWriter) })

	Setup("json", false, false)
	PrintInfo("uploading %s", "x")
	PrintSuccess("done")
	PrintWarning("careful")

	if buf.Len() != 0 {
		t.Fatalf("expected no data-writer output from status messages, got %q", buf.String())
	}

	if err := Print(map[string]string{"ok": "yes"}); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Fatalf("expected JSON document on data writer, got %q", buf.String())
	}
}
