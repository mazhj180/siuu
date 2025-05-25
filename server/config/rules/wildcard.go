package rules

import "siuu/tunnel/routing/rule"

type WildcardRule struct {
	rule.BaseRule
}

func (r *WildcardRule) Match(host string) (string, bool) {
	rl := len(r.Rule[1:])
	hl := len(host)
	return r.Target, hl >= rl && host[hl-rl:] == r.Rule[1:]
}
