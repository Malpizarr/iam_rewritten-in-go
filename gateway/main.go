package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()

	authProxy := newReverseProxy("http://localhost:8080")
	mux.Handle("/auth/", authProxy)

	graphqlProxy := newReverseProxy("http://localhost:8083")
	mux.Handle("/graphql", graphqlProxy)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:3000", "http://localhost:3000", "https://bookfront-r6l1.onrender.com", "https://bookfront-delta.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	log.Println("El servidor se está ejecutando en http://localhost:8085")
	log.Fatal(http.ListenAndServe(":8085", handler))
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.Director = func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", url.Host)
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.URL.Path = url.Path + req.URL.Path
	}

	return proxy
}
