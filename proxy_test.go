package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFetchVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/example.com/!module/@v/list" {
			t.Fatalf("expected request to /example.com/!module/@v/list, got %s", r.URL.Path)
		}
		fmt.Fprintln(w, "v0.0.0-20170915032832-14c0d48ead0c")
		fmt.Fprintln(w, "v0.51.0-alpha.0")
		fmt.Fprintln(w, "v1.0.0")
		fmt.Fprintln(w, "v1.5.1-0.20250403130103-3d3abc24416a")
		fmt.Fprintln(w, "v1.1.0-alpha1")
		fmt.Fprintln(w, "v2.0.0")
	}))
	defer server.Close()

	tests := []struct {
		version  string
		expected []string
	}{
		{
			"v0.0.0-20100915032832-14c0d48ead0c", []string{
				"v2.0.0",
				"v1.5.1-0.20250403130103-3d3abc24416a",
				"v1.1.0-alpha1",
				"v1.0.0",
				"v0.51.0-alpha.0",
				"v0.0.0-20170915032832-14c0d48ead0c",
			},
		},
		{
			"v0.0.1", []string{
				"v2.0.0",
				"v1.0.0",
			},
		},
		{
			"v1.1.0-alpha1", []string{
				"v2.0.0",
				"v1.5.1-0.20250403130103-3d3abc24416a",
			},
		},
		{
			"v1.0.0", []string{
				"v2.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			proxy := NewGoProxy(server.URL)
			versionsMod, err := proxy.FetchVersions("example.com/Module", tt.version)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			var versions []string
			for _, v := range versionsMod {
				versions = append(versions, v.Version)
			}

			if diff := cmp.Diff(tt.expected, versions); diff != "" {
				t.Errorf("unexpected versions (-want +got):\n%s", diff)
			}
		})
	}
}
