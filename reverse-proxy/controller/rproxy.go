package controller

import (
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	"net/http"
	"net/http/httputil"
	"strings"
)

type ReverseProxy struct {
	ing networkingv1.Ingress
	rp *httputil.ReverseProxy
}

func NewIngressHandler(ing networkingv1.Ingress) (http.Handler, error) {
	if len(ing.Spec.Rules) == 0 { // TODO: remove it
		return http.NewServeMux(), nil
	}

	director := func(req *http.Request) {
		var defTargetHost string
		if ing.Spec.DefaultBackend != nil {
			defTargetHost =  urlFromIngressServiceBackend(ing.Spec.DefaultBackend.Service)
		}
		for _, r := range ing.Spec.Rules {
			if r.Host != req.Host {
				continue
			}

			for _, p := range r.IngressRuleValue.HTTP.Paths {
				if (strings.HasPrefix(req.URL.Path, p.Path)) {
					defTargetHost = urlFromIngressServiceBackend(p.Backend.Service)
					break
				}
			}
		}

		req.URL.Scheme = "http"
		req.URL.Host = defTargetHost
		/*if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}*/
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}, nil
}

func (r *ReverseProxy) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	r.rp.ServeHTTP(writer, req)
}

func urlFromIngressServiceBackend(s *networkingv1.IngressServiceBackend) string {
	return fmt.Sprintf("%s:%v", s.Name, s.Port.Number)
}
