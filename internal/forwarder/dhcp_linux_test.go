package forwarder

import (
	"reflect"
	"testing"
)

func Test_parse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantErr bool
	}{
		{
			name: "single valid IP",
			content: `
				# This is a comment
				nameserver 8.8.8.8
			`,
			want:    []string{"8.8.8.8"},
			wantErr: false,
		},
		{
			name: "multiple valid IPs",
			content: `
				# Multiple nameservers
				nameserver 8.8.8.8
				nameserver 8.8.4.4
				nameserver 1.1.1.1
			`,
			want:    []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
			wantErr: false,
		},
		{
			name: "invalid IP",
			content: `
				# Invalid IP
				nameserver 256.256.256.256
			`,
			wantErr: true,
		},
		{
			name: "mixed valid and invalid IPs",
			content: `
				nameserver 8.8.8.8
				nameserver invalid.ip
				nameserver 8.8.4.4
			`,
			want:    []string{"8.8.8.8", "8.8.4.4"},
			wantErr: false,
		},
		{
			name:    "empty file",
			content: ``,
			wantErr: true,
		},
		{
			name: "only comments",
			content: `
				# Just a comment
				# Another comment
			`,
			wantErr: true,
		},
		{
			name: "malformed lines",
			content: `
				nameserver
				nameserver 8.8.8.8
			`,
			want:    []string{"8.8.8.8"},
			wantErr: false,
		},
		{
			name: "with extra whitespace",
			content: `
				   nameserver    8.8.8.8
				   nameserver    8.8.4.4
			`,
			want:    []string{"8.8.8.8", "8.8.4.4"},
			wantErr: false,
		},
		{
			name: "IPv6 addresses",
			content: `
				nameserver 2001:4860:4860::8888
				nameserver 2001:4860:4860::8844
			`,
			want:    []string{"2001:4860:4860::8888", "2001:4860:4860::8844"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse([]byte(tt.content))

			if tt.wantErr {
				if err == nil {
					t.Errorf("parse() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
