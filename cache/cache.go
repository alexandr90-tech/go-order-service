package cache

import (
	"encoding/json"
	"net/http"
	"sync"

	"go-order-service/db"

	"github.com/gorilla/mux"
)

type Cache struct {
	sync.RWMutex
	Orders map[string]db.Order
}

func NewCache() *Cache {
	return &Cache{Orders: make(map[string]db.Order)}
}

func (c *Cache) LoadFromDB() error {
	orders, err := db.GetAllOrders()
	if err != nil {
		return err
	}
	c.Lock()
	for _, o := range orders {
		c.Orders[o.OrderUID] = o
	}
	c.Unlock()
	return nil
}

func (c *Cache) Set(orderUID string, order db.Order) {
	c.Lock()
	defer c.Unlock()
	c.Orders[orderUID] = order
}

func (c *Cache) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	c.RLock()
	order, ok := c.Orders[id]
	c.RUnlock()

	if !ok {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (c *Cache) GetOrder(orderUID string) (db.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	order, ok := c.Orders[orderUID]
	return order, ok
}
