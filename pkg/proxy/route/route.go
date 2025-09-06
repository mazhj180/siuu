package route

import (
	"errors"
	"siuu/pkg/proxy/client"
	"strings"
	"sync"
)

type Router interface {
	Name() string // return the name of the router

	Route(string) (client.ProxyClient, string, bool) // route the request to the proxy

	Initialize([]RouteRule, []client.ProxyClient, map[string]string, client.ProxyClient) error // initialize the route table and default outlet

	SetDefaultOutletIfExists(string) error // set the default outlet if exists

	SetProxy(string, client.ProxyClient) error // set a proxy to the route table

	SetMapping(string, string) error // set a mapping to the route table

	GetOriginalInfo() ([]client.ProxyClient, map[string]string, []RouteRule, client.ProxyClient) // get the original info of the router
}

// r is the builtin router
type r struct {
	mu            sync.RWMutex
	defaultOutlet client.ProxyClient
	clients       map[string]client.ProxyClient
	mappings      map[string]string // mapping of proxy alias to proxy name
	routeTable    *node

	originalRules   []RouteRule          // original rules
	originalClients []client.ProxyClient // original clients
}

// NewRouter creates a new builtin router
func NewRouter() Router {
	return &r{
		clients: make(map[string]client.ProxyClient),
		routeTable: &node{
			children:         make(map[string]*node),
			wildcardChildren: make(map[string]*node),
		},
	}
}

func (r *r) Name() string {
	return "builtin"
}

func (r *r) Route(host string) (client.ProxyClient, string, bool) {

	var isDefaultOutlet bool
	if node, exists := r.routeTable.children[host]; exists && node.typ == special {
		return r.clients[node.proxyName], node.segment, isDefaultOutlet
	}

	segments := strings.Split(host, ".")

	ruleSegments := make([]string, 0, len(segments))
	prxName := r.route(r.routeTable, segments, &ruleSegments)

	rule := strings.Join(ruleSegments, ".")
	if prxName == "direct" {
		return nil, rule, isDefaultOutlet
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if prx, exists := r.clients[prxName]; !exists {
		if prx, exists = r.clients[r.mappings[prxName]]; exists {
			return prx, rule, isDefaultOutlet
		}
	} else {
		return prx, rule, isDefaultOutlet
	}

	isDefaultOutlet = true

	return r.defaultOutlet, "", isDefaultOutlet
}

func (r *r) route(n *node, segments []string, ruleSegments *[]string) string {
	var l int
	if l = len(segments); l == 0 {
		return n.proxyName
	}

	segment := segments[l-1]
	remaining := segments[:l-1]

	if child, exists := n.children[segment]; exists {
		result := r.route(child, remaining, ruleSegments)
		*ruleSegments = append(*ruleSegments, child.segment)
		return result
	}

	for k, v := range n.wildcardChildren {
		sl, kl := len(segment), len(k)
		if (sl >= kl || sl+1 == kl) && strings.HasSuffix(segment, k[1:kl]) {
			*ruleSegments = append(*ruleSegments, v.segment)
			return v.proxyName
		}
	}

	return ""
}

func (r *r) Initialize(rules []RouteRule, proxies []client.ProxyClient, mappings map[string]string, defaultOutlet client.ProxyClient) error {

	if defaultOutlet != nil {
		r.defaultOutlet = defaultOutlet
	}
	r.mappings = mappings

	for _, prx := range proxies {
		r.clients[prx.Name()] = prx
	}
	r.originalClients = proxies

	r.originalRules = rules

	for _, rule := range rules {

		var exists bool
		var prxName string
		if _, exists = r.clients[rule.value]; !exists {
			_, exists = r.clients[mappings[rule.value]]
		}

		if !exists && rule.value != "direct" {
			continue
		} else {
			prxName = rule.value
		}

		segments := strings.Split(rule.key, ".")

		switch rule.typ {
		case "domain", "ip":
			r.addRuleNode(r.routeTable, segments, prxName)
		default:
			r.routeTable.children[rule.key] = &node{
				segment:   rule.key,
				typ:       special,
				proxyName: prxName,
			}
		}
	}

	return nil
}

func (r *r) addRuleNode(n *node, segments []string, name string) {

	l := len(segments)

	if l == 0 {
		n.proxyName = name
		return
	}

	segment := segments[l-1]
	remaining := segments[:l-1]

	nn := &node{
		segment: segment,
	}

	if strings.HasPrefix(segment, "*") {
		nn.typ = wildcard
		if n.wildcardChildren[segment] == nil {
			n.wildcardChildren[segment] = nn
		}
		r.addRuleNode(n.wildcardChildren[segment], nil, name)
		return
	}

	nn.typ = static
	nn.children = make(map[string]*node)
	nn.wildcardChildren = make(map[string]*node)

	if n.children[segment] == nil {
		n.children[segment] = nn
	}

	r.addRuleNode(n.children[segment], remaining, name)

}

func (r *r) SetDefaultOutletIfExists(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "direct" || name == "none" {
		r.defaultOutlet = nil
		return nil
	}

	if _, exists := r.clients[name]; !exists {
		return errors.New("proxy not found")
	}

	r.defaultOutlet = r.clients[name]
	return nil
}

func (r *r) SetProxy(name string, proxy client.ProxyClient) error {

	if proxy == nil {
		return errors.New("proxy is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if name != proxy.Name() && name != "direct" {
		r.mappings[name] = proxy.Name()
		return nil
	}

	r.clients[name] = proxy
	return nil
}

func (r *r) SetMapping(name string, proxyName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[proxyName]; !exists {
		return errors.New("proxy not found")
	}

	r.mappings[name] = proxyName
	return nil
}

func (r *r) GetOriginalInfo() ([]client.ProxyClient, map[string]string, []RouteRule, client.ProxyClient) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]client.ProxyClient, len(r.originalClients))
	copy(clients, r.originalClients)

	mappings := make(map[string]string, len(r.mappings))
	for k, v := range r.mappings {
		mappings[k] = v
	}

	copyRules := make([]RouteRule, len(r.originalRules))
	copy(copyRules, r.originalRules)

	return clients, mappings, copyRules, r.defaultOutlet
}

type RouteRule struct {
	typ   string
	key   string
	value string
}

func NewRouteRule(typ, key, value string) RouteRule {
	return RouteRule{
		typ:   typ,
		key:   key,
		value: value,
	}
}

func (rr *RouteRule) Type() string {
	return rr.typ
}

func (rr *RouteRule) Key() string {
	return rr.key
}

func (rr *RouteRule) Value() string {
	return rr.value
}

type nodeType uint8

const (
	static nodeType = iota
	wildcard
	special
)

type node struct {
	children         map[string]*node
	wildcardChildren map[string]*node

	typ nodeType

	segment   string
	proxyName string
}
