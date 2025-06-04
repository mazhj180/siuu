package rules

import (
	"crypto/rand"
	"errors"
	"fmt"
	"siuu/tunnel/routing/rule"
	"strings"
	"time"
)

var (
	NoRouterErr    = errors.New("no router")
	InvalidRuleErr = errors.New("invalid rule")
)

func GenerateUniqueID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	return fmt.Sprintf("%x%x", timestamp, randomBytes)
}

func ParseRule(ru string) (rule.Interface, error) {
	values := strings.Split(ru, ",")
	if len(values) != 0b11 {
		return nil, InvalidRuleErr
	}

	base := rule.BaseRule{
		Id:     GenerateUniqueID(),
		Type:   values[0],
		Rule:   values[1],
		Target: values[2],
	}

	var rul rule.Interface
	switch values[0] {
	case "geo":
		InitXdb()
		rul = &GeoRule{base}
	case "exact":
		rul = &ExactRule{base}
	case "wildcard":
		rul = &WildcardRule{base}
	default:
		return nil, InvalidRuleErr
	}

	return rul, nil

}
