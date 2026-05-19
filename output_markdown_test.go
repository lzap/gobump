package main

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestOutputMarkdown(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputMarkdown(&buf)

	out.Begin("Test Begin")
	out.Header("Test Header")
	out.BeginPreformatted("Test Preformatted")
	out.Println("Text")
	out.Error("Error")
	out.EndPreformatted()
	out.BeginPreformatted("Test Preformatted")
	out.Println("Text")
	out.EndPreformattedCond(false)
	out.End("Test End")

	expected := `Test Begin

### Test Header

<details><summary>Test Preformatted</summary>

` + "```" + `
Text
Error
` + "```" + `
</details>
Test End

:pretzel: *Created with [gobump](https://github.com/lzap/gobump) (HEAD)* :pretzel:
`

	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("OutputMarkdown mismatch (-want +got):\n%s", diff)
	}
}

func TestOutputMarkdownPrintSummary(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputMarkdown(&buf)

	out.PrintSummary([]Result{
		{
			ModulePath:    "example.com/mod",
			Success:       true,
			VersionBefore: "v1.0.0",
			VersionAfter:  "v2.0.0",
		},
		{
			ModulePath:    "example.com/unchanged",
			Success:       true,
			VersionBefore: "v1.0.0",
			VersionAfter:  "v1.0.0",
		},
	})

	expected := `
## Summary

| Module | Status | Version |
| --- | --- | --- |
| example.com/mod | U | v1.0.0 > v2.0.0 |
| example.com/unchanged | - | v1.0.0 > v1.0.0 |

Status: **U** updated, **E** error, **X** excluded, **N** no newer versions on module proxy, **-** unchanged.
`

	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("PrintSummary mismatch (-want +got):\n%s", diff)
	}
}
