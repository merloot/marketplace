package httpapi

import (
	"net/http"
)

func NewRouter(stockHandler *StockHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /stock", stockHandler.Add)

	var handler http.Handler = mux
	handler = Recovery(stockHandler.log)(handler)
	handler = Logging(stockHandler.log)(handler)
	return mux
}
