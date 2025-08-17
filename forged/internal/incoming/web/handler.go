package web

import "net/http"

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
