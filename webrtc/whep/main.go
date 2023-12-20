package main

import (
	"io"
	"log"
	"net/http"
	"path"
	"strings"
)

const (
	HTTP_ADDR = ":8082"
	WHEP_EXT  = ".whep"
)

type whepHandler struct {
	httpAddr string
}

func (h *whepHandler) Init() {

}

func (h *whepHandler) createWhepClient(path string, offer []byte) (answer []byte, err error) {
	return nil, nil
}

func (h *whepHandler) deleteWhepClient(path string) error {
	return nil
}

func (h *whepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Ext(r.URL.Path) != WHEP_EXT {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		scheme := "http://"
		if r.TLS != nil {
			scheme = "https://"
		}
		offer, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		answer, err := h.createWhepClient(r.URL.Path, offer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", strings.Join([]string{scheme, r.Host, r.URL.Path}, ""))
		w.Header().Set("Content-Type", "application/sdp")
		w.WriteHeader(http.StatusCreated)
		w.Write(answer)
		return
	case http.MethodDelete:
		if err := h.deleteWhepClient(r.URL.Path); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,DELETE,OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {

	h := &whepHandler{
		httpAddr: HTTP_ADDR,
	}
	h.Init()
	log.Println("whep demo running", h.httpAddr)
	log.Fatal(http.ListenAndServe(h.httpAddr, h))
}
