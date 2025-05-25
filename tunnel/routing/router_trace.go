package routing

import "fmt"

type TraceLevel = byte

const (
	ALL TraceLevel = iota
	PREFERRED
)

type Tracer struct{}

func (t *Tracer) Trace(content string, level TraceLevel, traces *[]Trace, isTrace bool) {

	if !isTrace {
		return
	}

	trace := Trace{Level: level, Content: content}
	*traces = append(*traces, trace)
}

type Trace struct {
	Level   TraceLevel // for filtering
	Content string
}

func (t Trace) String() string {
	return fmt.Sprintf("[%s]", t.Content)
}
