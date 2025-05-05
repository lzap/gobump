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
` + "```" + `
</details>
Test End
:pretzel: *Created with [gobump](https://github.com/lzap/gobump) (HEAD)* :pretzel:
`

	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("OutputMarkdown mismatch (-want +got):\n%s", diff)
	}
}
