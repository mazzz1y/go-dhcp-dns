package main

import (
	"testing"
)

func TestParseDNSServers(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []string
		wantErr bool
	}{
		{
			name: "one valid DNS server",
			input: []byte(`
				domain_name_server (ip_multicast): {192.168.1.1}
			`),
			want:    []string{"192.168.1.1"},
			wantErr: false,
		},
		{
			name: "two valid DNS servers",
			input: []byte(`
				domain_name_server (ip_multicast): {192.168.1.1, 8.8.8.8}
			`),
			want:    []string{"192.168.1.1", "8.8.8.8"},
			wantErr: false,
		},
		{
			name: "three valid DNS servers",
			input: []byte(`
				domain_name_server (ip_multicast): {192.168.1.1, 8.8.8.8, 8.8.4.4}
			`),
			want:    []string{"192.168.1.1", "8.8.8.8", "8.8.4.4"},
			wantErr: false,
		},
		{
			name: "zero valid DNS servers",
			input: []byte(`
				domain_name_server (ip_multicast): {}
			`),
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid DNS server format",
			input: []byte(`
				domain_name_server (ip_multicast): 192.168.1.1, 8.8.8.8
			`),
			want:    nil,
			wantErr: true,
		},
		{
			name: "no DNS server section",
			input: []byte(`
				some_other_info: {192.168.1.1, 8.8.8.8}
			`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDNSServers(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDNSServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("parseDNSServers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
