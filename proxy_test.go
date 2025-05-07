package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/mod/module"
)

func TestFetchVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/example.com/@v/list" {
			t.Fatalf("expected request to /example.com/@v/list, got %s", r.URL.Path)
		}
		fmt.Fprintln(w, "v2.0.0")
		fmt.Fprintln(w, "v1.0.0")
		fmt.Fprintln(w, "v1.1.0")
	}))
	defer server.Close()

	proxy := NewGoProxy(server.URL)
	versions, err := proxy.FetchVersions("example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []module.Version{
		{Path: "example.com", Version: "v2.0.0"},
		{Path: "example.com", Version: "v1.1.0"},
		{Path: "example.com", Version: "v1.0.0"},
	}

	if diff := cmp.Diff(expected, versions); diff != "" {
		t.Errorf("unexpected versions (-want +got):\n%s", diff)
	}
}
