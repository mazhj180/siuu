package handle

import (
	"encoding/json"
	"io"
	"net/http"
	"siuu/server/store"
	"strings"
)

func RegisterProxyHandle(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/add", addProxy)
	mux.HandleFunc(prefix+"/remove", removeProxy)
	mux.HandleFunc(prefix+"/get", getProxies)
	mux.HandleFunc(prefix+"/set", setDefaultProxy)
}

func addProxy(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var proxies []string
	if err = json.Unmarshal(body, &proxies); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = store.AddProxies(proxies...); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func removeProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("proxies") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := query.Get("proxies")
	var names []string
	for _, name := range strings.Split(params, ",") {
		names = append(names, name)
	}
	store.RemoveProxies(names...)
	w.WriteHeader(http.StatusOK)
}

func getProxies(w http.ResponseWriter, r *http.Request) {
	proxies := store.GetProxies()
	data, err := json.Marshal(proxies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func setDefaultProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if !query.Has("proxy") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := store.SetSelectedProxy(query.Get("proxy"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func test(w http.ResponseWriter, r *http.Request) {

}
