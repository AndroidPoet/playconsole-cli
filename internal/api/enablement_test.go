package api

import (
	"strings"
	"testing"
)

func TestHandleAPIEnablementReturnsURLWhenNonInteractive(t *testing.T) {
	apiErr := &APIEnablementError{
		APITitle:      "Google Play Android Developer API",
		ActivationURL: "https://console.cloud.google.com/apis/api/androidpublisher.googleapis.com/overview?project=123456",
	}

	err := HandleAPIEnablement(apiErr)
	if err == nil {
		t.Fatal("expected non-interactive enablement error")
	}

	message := err.Error()
	if !strings.Contains(message, apiErr.APITitle) {
		t.Fatalf("expected API title in error, got %q", message)
	}
	if !strings.Contains(message, apiErr.ActivationURL) {
		t.Fatalf("expected activation URL in error, got %q", message)
	}
}
