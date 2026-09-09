package auth

import (
	"strings"
	"testing"
)

func TestValidateServiceAccountJSONAcceptsServiceAccountKey(t *testing.T) {
	data := []byte(`{
  "type": "service_account",
  "project_id": "demo",
  "client_email": "bot@demo.iam.gserviceaccount.com",
  "private_key": "-----BEGIN PRIVATE KEY-----\nabc\n-----END PRIVATE KEY-----\n"
}`)
	if err := validateServiceAccountJSON(data); err != nil {
		t.Fatalf("expected valid key, got %v", err)
	}
}

func TestValidateServiceAccountJSONRejectsOtherCredentialTypes(t *testing.T) {
	cases := map[string]string{
		"oauth client":   `{"type":"authorized_user","client_id":"x","client_secret":"y"}`,
		"missing fields": `{"type":"service_account"}`,
		"not json":       `eyJ0eXBlIjoic2VydmljZV9hY2NvdW50In0=`,
	}
	for name, body := range cases {
		if err := validateServiceAccountJSON([]byte(body)); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestValidateServiceAccountJSONErrorMentionsType(t *testing.T) {
	err := validateServiceAccountJSON([]byte(`{"type":"authorized_user","client_email":"a","private_key":"b"}`))
	if err == nil || !strings.Contains(err.Error(), "authorized_user") {
		t.Fatalf("expected error naming the wrong type, got %v", err)
	}
}
