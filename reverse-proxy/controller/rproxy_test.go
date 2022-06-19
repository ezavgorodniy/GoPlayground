package controller_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"cookpad.com/global-sre/tech-test/controller"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const singleBackend = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          service:
            name: %s
            port:
              number: %s`

const pathBasedRouting = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: "/foo"
        backend:
          service:
            name: %s
            port:
              number: %s
      - path: "/bar"
        backend:
          service:
            name: %s
            port:
              number: %s`

const hostBasedRouting = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          service:
            name: %s
            port:
              number: %s
  - host: bar.foo.com
    http:
      paths:
      - backend:
          service:
            name: %s
            port:
              number: %s`

const emptyIngress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test`

const defaultBackend = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test
spec:
  defaultBackend:
    service:
      name: %s
      port:
        number: %s
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          service:
            name: %s
            port:
              number: %s`

type request struct {
	host string
	path string
}

type expectation struct {
	code int
	body string
}

func TestReverseProxy(t *testing.T) {
	backend1, name1, port1 := makeBackendServer("backend1")
	defer backend1.Close()
	backend2, name2, port2 := makeBackendServer("backend2")
	defer backend2.Close()

	cases := map[string]struct {
		ingressDefintion string
		request          request
		expected         expectation
	}{
		"single backend routing 1": {
			ingressDefintion: fmt.Sprintf(singleBackend, name1, port1),
			request: request{
				host: "foo.bar.com",
				path: "/",
			},
			expected: expectation{
				code: 200,
				body: "backend1: /",
			},
		},
		"single backend routing 2": {
			ingressDefintion: fmt.Sprintf(singleBackend, name1, port1),
			request: request{
				host: "foo.bar.com",
				path: "/foo",
			},
			expected: expectation{
				code: 200,
				body: "backend1: /foo",
			},
		},
		"path based routing 1": {
			ingressDefintion: fmt.Sprintf(pathBasedRouting, name1, port1, name2, port2),
			request: request{
				host: "foo.bar.com",
				path: "/foo",
			},
			expected: expectation{
				code: 200,
				body: "backend1: /foo",
			},
		},
		"path based routing 2": {
			ingressDefintion: fmt.Sprintf(pathBasedRouting, name1, port1, name2, port2),
			request: request{
				host: "foo.bar.com",
				path: "/bar",
			},
			expected: expectation{
				code: 200,
				body: "backend2: /bar",
			},
		},
		"host based routing 1": {
			ingressDefintion: fmt.Sprintf(hostBasedRouting, name1, port1, name2, port2),
			request: request{
				host: "foo.bar.com",
			},
			expected: expectation{
				code: 200,
				body: "backend1: /",
			},
		},
		"host based routing 2": {
			ingressDefintion: fmt.Sprintf(hostBasedRouting, name1, port1, name2, port2),
			request: request{
				host: "bar.foo.com",
			},
			expected: expectation{
				code: 200,
				body: "backend2: /",
			},
		},
		"host based routing 3": {
			ingressDefintion: fmt.Sprintf(hostBasedRouting, name1, port1, name2, port2),
			request: request{
				host: "example.org",
				path: "/",
			},
			expected: expectation{
				code: 404,
				body: "404 page not found\n",
			},
		},
		"no backends": {
			ingressDefintion: emptyIngress,
			expected: expectation{
				code: 404,
				body: "404 page not found\n",
			},
		},
		"default backend 1": {
			ingressDefintion: fmt.Sprintf(defaultBackend, name1, port1, name2, port2),
			request: request{
				host: "anything.com",
				path: "/whatever",
			},
			expected: expectation{
				code: 200,
				body: "backend1: /whatever",
			},
		},
		"default backend 2": {
			ingressDefintion: fmt.Sprintf(defaultBackend, name1, port1, name2, port2),
			request: request{
				host: "foo.bar.com",
				path: "/whatever",
			},
			expected: expectation{
				code: 200,
				body: "backend2: /whatever",
			},
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ingress, err := parseIngressYaml(tt.ingressDefintion)
			if err != nil {
				t.Error(err)
			}

			handler, err := controller.NewIngressHandler(ingress)
			if err != nil {
				t.Error(err)
			}

			reverseProxy := httptest.NewServer(handler)
			defer reverseProxy.Close()

			client := reverseProxy.Client()
			getReq, _ := http.NewRequest("GET", reverseProxy.URL, nil)
			getReq.Host = tt.request.host
			getReq.URL.Path = tt.request.path
			getReq.Close = true
			res, err := client.Do(getReq)

			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.expected.code, res.StatusCode)

			body, _ := ioutil.ReadAll(res.Body)
			assert.Equal(t, tt.expected.body, string(body))
		})
	}
}

func makeBackendServer(body string) (*httptest.Server, string, string) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(body + ": " + r.URL.Path))
	}))
	backendURL, _ := url.Parse(backend.URL)
	host := regexp.MustCompile(`(\d|.+):(\d+)`).FindStringSubmatch(backendURL.Host)
	return backend, host[1], host[2]
}

func parseIngressYaml(yaml string) (networkingv1.Ingress, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		return networkingv1.Ingress{}, err
	}
	ingress, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return networkingv1.Ingress{}, errors.New("Parsed object is not an Ingress")
	}
	return *ingress, nil
}
