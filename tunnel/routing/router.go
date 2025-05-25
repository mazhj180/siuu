package routing

import (
	"errors"
	"fmt"
	"siuu/tunnel/routing/rule"
	"sync"
)

var (
	NotFoundRuleErr   = errors.New("not found rule")
	MissAnyRulesErr   = errors.New("miss any rules")
	HitIgnoredRuleErr = errors.New("hit ignored rule")

	IgnoreFlag = "-1"
)

type Router interface {
	Name() string
	Route(string, bool) (target string, traces []Trace, err error)
	AddRule(rule.Interface)
	IgnoreRule(string) error
	GetRule(string) (rule.Interface, error)
}

type BasicRouter struct {
	sync.RWMutex

	Rules       map[string]rule.Interface // for searching
	RuleIds     []string                  // for sorting
	IgnoreRules map[string]struct{}

	Tracer
}

func (r *BasicRouter) Name() string {
	return "basic"
}

func (r *BasicRouter) Route(dst string, trace bool) (string, []Trace, error) {

	r.RLock()
	defer r.RUnlock()

	var target, dstId string
	var traces []Trace

	r.Trace(fmt.Sprintf("router [%s] dst [%s]", r.Name(), dst), PREFERRED, &traces, trace)

	for _, id := range r.RuleIds {
		ru := r.Rules[id]

		r.Trace(fmt.Sprintf("checking rule [%s]", ru), ALL, &traces, trace)

		if _, ok := r.IgnoreRules[id]; ok {
			dstId = IgnoreFlag
			r.Trace(fmt.Sprintf("hit ignored rule [%s] skipped", ru), ALL, &traces, trace)
			continue
		}

		if tar, ok := ru.Match(dst); ok {
			target, dstId = tar, id
			r.Trace(fmt.Sprintf("matched successfully hit rule [%s]", ru), PREFERRED, &traces, trace)
			break
		}

		r.Trace(fmt.Sprintf("miss rule [%s]", ru), ALL, &traces, trace)
	}

	if dstId == "" {
		r.Trace("miss any rules", PREFERRED, &traces, trace)
		return target, traces, MissAnyRulesErr
	}

	if dstId == IgnoreFlag {
		r.Trace("unmatched", PREFERRED, &traces, trace)
		return target, traces, HitIgnoredRuleErr
	}

	return target, traces, nil
}

func (r *BasicRouter) AddRule(ru rule.Interface) {
	r.Lock()
	defer r.Unlock()

	r.Rules[ru.ID()] = ru
	r.RuleIds = append(r.RuleIds, ru.ID())
}

func (r *BasicRouter) IgnoreRule(id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.Rules[id]; !ok {
		return NotFoundRuleErr
	}

	r.IgnoreRules[id] = struct{}{}
	return nil
}

func (r *BasicRouter) GetRule(id string) (rule.Interface, error) {

	r.RLock()
	defer r.RUnlock()

	var ru rule.Interface
	var ok bool
	if ru, ok = r.Rules[id]; !ok {
		return ru, NotFoundRuleErr
	}

	if _, ok = r.IgnoreRules[id]; ok {
		return ru, HitIgnoredRuleErr
	}

	return ru, nil
}
