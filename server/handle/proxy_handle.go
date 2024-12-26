package handle

import (
	"encoding/json"
	"io"
	"net/http"
	"siu/tunnel/proxy"
	"strings"
)

func RegisterProxyHandle(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/add", addProxy)
	mux.HandleFunc(prefix+"/remove", removeProxy)
	mux.HandleFunc(prefix+"/get", getProxies)
}

func addProxy(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var pws []proxy.ProxyWrapper
	if err = json.Unmarshal(body, &pws); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var proxies []proxy.Proxy
	for _, pw := range pws {
		proxies = append(proxies, pw.Value)
	}

	if err = proxy.AddProxies(proxies...); err != nil {
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
	proxy.RemoveProxies(names...)
	w.WriteHeader(http.StatusOK)
}

func getProxies(w http.ResponseWriter, r *http.Request) {
	proxies := proxy.GetProxies()
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
