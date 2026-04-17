package provider

import "testing"

func TestValidateAPIBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"default production URL", defaultAPIBaseURL, false},
		{"custom https host", "https://api.example.com/v1", false},
		{"http scheme rejected", "http://api.sporkops.com/v1", true},
		{"ftp scheme rejected", "ftp://api.sporkops.com", true},
		{"literal loopback v4 rejected", "https://127.0.0.1:8080/v1", true},
		{"literal loopback v6 rejected", "https://[::1]/v1", true},
		{"literal private 10.x rejected", "https://10.0.0.5/v1", true},
		{"literal private 192.168.x rejected", "https://192.168.1.1/v1", true},
		{"literal link-local rejected", "https://169.254.169.254/v1", true},
		{"literal unspecified rejected", "https://0.0.0.0/v1", true},
		{"localhost hostname rejected", "https://localhost/v1", true},
		{"google metadata hostname rejected", "https://metadata.google.internal/v1", true},
		{"empty host rejected", "https:///v1", true},
		{"garbage rejected", "://not a url", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAPIBaseURL(tc.url)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q, got nil", tc.url)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.url, err)
			}
		})
	}
}
