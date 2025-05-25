package rules

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/spf13/viper"
	"io"
	"os"
	"siuu/logger"
	"siuu/server/config/constant"
	"siuu/tunnel/routing"
	"siuu/tunnel/routing/rule"
	"siuu/util"
	"strings"
	"time"
)

var (
	NoRouterErr = errors.New("no router")
)

func GenerateUniqueID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	return fmt.Sprintf("%x%x", timestamp, randomBytes)
}

func LoadRules(router routing.Router) error {

	if router == nil {
		return NoRouterErr
	}

	paths := util.GetConfigSlice(constant.RuleRoutePath)
	xdbp = util.GetConfig[string](constant.RuleRouteXdbPath)
	xdbp = util.ExpandHomePath(xdbp)
	xdbb, _ = xdb.LoadVectorIndexFromFile(xdbp)

	v := viper.New()
	v.SetConfigType("toml")

	for _, path := range paths {
		path = util.ExpandHomePath(path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		fin, err := os.OpenFile(path, os.O_RDONLY, 0666)
		if err != nil {
			continue
		}
		defer fin.Close()

		hasher := sha256.New()
		_, err = io.Copy(hasher, fin)
		if err != nil {
			logger.SWarn("failed to initialize router [%s]", path)
			continue
		}

		signature := fmt.Sprintf("%xroute", hasher.Sum(nil))
		if s, ok := constant.Signature[path]; ok && s == signature {
			continue
		}
		constant.Signature[path] = signature

		_, err = fin.Seek(0, io.SeekStart)
		if err != nil {
			logger.SWarn("failed to initialize router [%s]", path)
			continue
		}

		if err = v.ReadConfig(fin); err != nil {
			return err
		}

		rules := v.GetStringSlice("route.rules")
		for _, ru := range rules {
			values := strings.Split(ru, ",")
			if len(values) != 0b11 {
				continue
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
				rul = &GeoRule{base}
			case "exact":
				rul = &ExactRule{base}
			case "wildcard":
				rul = &WildcardRule{base}
			default:
				continue
			}

			router.AddRule(rul)
		}
	}

	return nil
}
