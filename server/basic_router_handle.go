package server

import (
	"encoding/json"
	"net/http"
	"siuu/tunnel/routing"
	"siuu/util/pinyin"
)

func init() {
	handlerMapping := make(map[string]http.HandlerFunc)

	handlerMapping["/alias/set"] = setAlias
	handlerMapping["/alias/get"] = getAliases
	srv.RegisterHandlerFunc(handlerMapping)
}

func getAliases(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	srv.RLock()
	defer srv.RUnlock()

	router := srv.Router

	router.RLock()
	defer router.RUnlock()

	var ok bool
	var br *routing.BasicRouter
	if br, ok = router.(*routing.BasicRouter); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var alias []string
	for k, _ := range br.ProxyAlias {
		alias = append(alias, k)
	}

	if query.Has("alias") {
		ali := query.Get("alias")
		alias = pinyin.FuzzyMatch(alias, ali)
	}

	aliases := make([]any, 0, len(alias))

	for _, v := range alias {
		aliasStruct := struct {
			Alias string `json:"alias"`
			Proxy string `json:"proxy"`
		}{
			Alias: v,
			Proxy: br.ProxyAlias[alias[0]],
		}
		aliases = append(aliases, aliasStruct)
	}

	data, err := json.Marshal(aliases)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func setAlias(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("alias") || !query.Has("proxy") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.RLock()
	defer srv.RUnlock()

	router, ok := srv.Router.(*routing.BasicRouter)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	router.RLock()
	defer router.RUnlock()

	router.ProxyAlias[query.Get("alias")] = query.Get("proxy")

	w.WriteHeader(http.StatusOK)
}
