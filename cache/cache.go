package cache

import (
	"encoding/json"
	"net/http"
	"sync"

	"go-order-service/db"

	"github.com/gorilla/mux"
)

// Cache — потокобезопасный кеш заказов
type Cache struct {
	sync.RWMutex
	Orders map[string]db.Order
}

// NewCache создаёт новый кеш
func NewCache() *Cache {
	return &Cache{Orders: make(map[string]db.Order)}
}

// LoadFromDB загружает все заказы из БД в кеш
func (c *Cache) LoadFromDB() error {
	orders, err := db.GetAllOrders()
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	for _, o := range orders {
		c.Orders[o.OrderUID] = o
	}
	return nil
}

// Set добавляет/обновляет заказ в кеше
func (c *Cache) Set(orderUID string, order db.Order) {
	c.Lock()
	defer c.Unlock()
	c.Orders[orderUID] = order
}

// GetOrderHandler — HTTP handler для получения заказа по UID
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
	_ = json.NewEncoder(w).Encode(order)
}

// GetOrder возвращает заказ по UID
func (c *Cache) GetOrder(orderUID string) (db.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	order, ok := c.Orders[orderUID]
	return order, ok
}
