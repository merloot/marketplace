package httpapi

import (
	"net/http"
)

func NewRouter(productHandler *ProductHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /products", productHandler.Create)

	return mux
}
