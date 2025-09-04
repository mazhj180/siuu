package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"siuu/api"
	"siuu/pkg/logger"
	"siuu/pkg/proxy/client"
	"siuu/pkg/proxy/route"
	"slices"
	"sync"
	"time"
)

func GetRouterHandlers(router route.Router, log *logger.Logger) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/api/router/clients": func(w http.ResponseWriter, r *http.Request) {
			clients, mappings, rules, defaultOutlet := router.GetOriginalInfo()

			clientsInfo := make([]api.ClientInfo, 0, len(clients))
			for _, client := range clients {
				clientsInfo = append(clientsInfo, api.ClientInfo{
					Name:        client.Name(),
					Type:        client.Type(),
					Server:      client.ServerHost(),
					Port:        client.ServerPort(),
					TrafficType: client.SupportTrafficType(),
				})
			}

			rulesInfo := make([]api.RuleInfo, 0, len(rules))
			for _, rule := range rules {
				rulesInfo = append(rulesInfo, api.RuleInfo{
					Typ:   rule.Type(),
					Key:   rule.Key(),
					Value: rule.Value(),
				})
			}

			var dp string
			if defaultOutlet == nil {
				dp = "none"
			} else {
				dp = defaultOutlet.Name()
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(api.Response[api.RouterInfo]{
				Code:    0,
				Message: "success",
				Data: api.RouterInfo{
					Clients:       clientsInfo,
					Mappings:      mappings,
					Rules:         rulesInfo,
					DefaultOutlet: dp,
				},
			}); err != nil {
				log.Error("JSON encode error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		},

		"/api/router/set/mappings": func(w http.ResponseWriter, r *http.Request) {
			var params api.ProxyParams
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				log.Error("JSON decode error: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if err := router.SetMapping(params.MappingName, params.ProxyName); err != nil {
				log.Error("Set mapping error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(api.Response[api.Data]{
				Code:    0,
				Message: "success",
				Data:    api.Data{},
			}); err != nil {
				log.Error("JSON encode error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		},

		"/api/router/set/default_outlet": func(w http.ResponseWriter, r *http.Request) {
			proxyName := r.URL.Query().Get("proxy_name")
			if proxyName == "" {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if err := router.SetDefaultOutletIfExists(proxyName); err != nil {
				log.Error("Set default outlet error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(api.Response[api.Data]{
				Code:    0,
				Message: "success",
				Data:    api.Data{},
			}); err != nil {
				log.Error("JSON encode error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		},

		"/api/router/proxy/ping": func(w http.ResponseWriter, r *http.Request) {
			var params api.PingProxyParam
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				log.Error("JSON decode error: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			url := params.TestUrl
			if url == "" {
				url = "https://www.github.com"
			}

			var tested []client.ProxyClient
			clients, _, _, _ := router.GetOriginalInfo()

			if len(params.Proxies) == 0 {
				tested = clients[:min(len(clients), 20)]
			} else {
				for _, client := range clients {
					if slices.Contains(params.Proxies, client.Name()) {
						tested = append(tested, client)
					}
				}
			}

			type result struct {
				Name  string
				Delay time.Duration
			}

			results := make(chan result, len(tested))
			var wg sync.WaitGroup

			for _, cli := range tested {
				wg.Add(1)
				go func(client client.ProxyClient) {
					defer wg.Done()
					start := time.Now()

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						results <- result{Name: client.Name(), Delay: 0}
						return
					}

					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancel()

					conn, err := client.Connect(ctx, "tcp", req.Host, 443)
					if err != nil {
						results <- result{Name: client.Name(), Delay: 0}
						return
					}
					defer conn.Close()

					results <- result{Name: client.Name(), Delay: time.Since(start)}

				}(cli)
			}

			// wait for all goroutines to complete, then close channel
			go func() {
				wg.Wait()
				close(results)
			}()

			resps := make([]api.PingResult, 0, len(tested))
			for res := range results {
				var delay string
				if res.Delay == 0 {
					delay = "timeout"
				} else {
					delay = fmt.Sprintf("%dms", res.Delay.Milliseconds())
				}

				resps = append(resps, api.PingResult{
					Name:  res.Name,
					Delay: delay,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(api.Response[[]api.PingResult]{
				Code:    0,
				Message: "success",
				Data:    resps,
			}); err != nil {
				log.Error("JSON encode error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		},
	}
}
