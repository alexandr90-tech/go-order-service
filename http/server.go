package http

import (
	"encoding/json"
	"go-order-service/cache"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	cache *cache.Cache
}

func NewServer(c *cache.Cache) *Server {
	return &Server{cache: c}
}

func (s *Server) Start(port string) {
	r := mux.NewRouter()

	r.HandleFunc("/order/{uid}", s.GetOrderHandler).Methods("GET")

	log.Printf("HTTP сервер запущен на :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func (s *Server) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	order, ok := s.cache.GetOrder(uid)
	if !ok {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Ошибка кодирования JSON: %v\n", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}
