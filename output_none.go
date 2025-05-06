package main

type OutputNone struct{}

var _ Output = (*OutputNone)(nil)

func (out *OutputNone) Begin(text ...any) {
}

func (out *OutputNone) End(text ...any) {
}

func (out *OutputNone) Header(text string) {
}

func (out *OutputNone) BeginPreformatted(text ...any) {
}

func (out *OutputNone) EndPreformatted(text ...any) {
}

func (out *OutputNone) EndPreformattedCond(render bool, text ...any) {
}

func (out *OutputNone) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func (out *OutputNone) Println(text ...string) {
}

func (out *OutputNone) Error(str ...string) {
}

func (out *OutputNone) Fatal(msg string, code ...int) {
}

func (out *OutputNone) PrintSummary(results []Result) {
}
