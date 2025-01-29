package forwarder

import (
	"reflect"
	"testing"
)

func Test_parse(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []string
		wantErr bool
	}{
		{
			name: "single valid IP",
			output: `
				domain_name_server {8.8.8.8}
			`,
			want:    []string{"8.8.8.8"},
			wantErr: false,
		},
		{
			name: "multiple valid IPs",
			output: `
				domain_name_server {8.8.8.8, 8.8.4.4, 1.1.1.1}
			`,
			want:    []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
			wantErr: false,
		},
		{
			name: "invalid IP",
			output: `
				domain_name_server {256.256.256.256}
			`,
			wantErr: true,
		},
		{
			name: "mixed valid and invalid IPs",
			output: `
				domain_name_server {8.8.8.8, invalid.ip, 8.8.4.4}
			`,
			want:    []string{"8.8.8.8", "8.8.4.4"},
			wantErr: false,
		},
		{
			name:    "empty file",
			output:  ``,
			wantErr: true,
		},
		{
			name: "no DNS section",
			output: `
				some_other_section {value}
				another_section {value}
			`,
			wantErr: true,
		},
		{
			name: "malformed DNS section - no brackets",
			output: `
				domain_name_server 8.8.8.8
			`,
			wantErr: true,
		},
		{
			name: "malformed DNS section - unclosed bracket",
			output: `
				domain_name_server {8.8.8.8
			`,
			wantErr: true,
		},
		{
			name: "with extra whitespace",
			output: `
				domain_name_server  {  8.8.8.8  ,  8.8.4.4  }
			`,
			want:    []string{"8.8.8.8", "8.8.4.4"},
			wantErr: false,
		},
		{
			name: "IPv6 addresses",
			output: `
				domain_name_server {2001:4860:4860::8888, 2001:4860:4860::8844}
			`,
			want:    []string{"2001:4860:4860::8888", "2001:4860:4860::8844"},
			wantErr: false,
		},
		{
			name: "multiple sections with DNS last",
			output: `
				some_other_section {value}
				another_section {value}
				domain_name_server {8.8.8.8, 8.8.4.4}
			`,
			want:    []string{"8.8.8.8", "8.8.4.4"},
			wantErr: false,
		},
		{
			name: "empty DNS section",
			output: `
				domain_name_server {}
			`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse([]byte(tt.output))

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
