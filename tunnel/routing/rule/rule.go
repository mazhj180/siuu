package rule

import "fmt"

type Interface interface {
	fmt.Stringer

	ID() string
	OriginalContent() (string, string)
	Match(host string) (string, bool)
}

type BaseRule struct {
	Id     string
	Type   string
	Rule   string
	Target string
}

func (r *BaseRule) Match(host string) (string, bool) {
	return r.Target, r.Rule == host
}

func (r *BaseRule) OriginalContent() (string, string) {
	return r.Rule, r.Target
}

func (r *BaseRule) ID() string {
	return r.Id
}

func (r *BaseRule) String() string {
	return fmt.Sprintf("[%s] %s -> %s", r.Type, r.Rule, r.Target)
}
