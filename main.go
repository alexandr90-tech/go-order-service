package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-order-service/cache"
	"go-order-service/db"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

func main() {
	// Загружаем конфиг
	cfg := LoadConfig()

	// Подключение к PostgreSQL
	if err := db.InitDB(cfg.GetDBConnString()); err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Инициализируем кеш
	c := cache.NewCache()
	if err := c.LoadFromDB(); err != nil {
		log.Printf("Не удалось загрузить кеш из БД: %v", err)
	}

	// Запускаем Kafka consumer в отдельной горутине
	go consumeKafka(cfg, c)

	// HTTP роуты
	r := mux.NewRouter()
	r.HandleFunc("/order/{id}", c.GetOrderHandler).Methods("GET")
	r.HandleFunc("/", serveWeb)

	// HTTP сервер
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
	}

	// Грациозное завершение
	go func() {
		log.Printf("HTTP сервер запущен на :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка HTTP сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Выключение сервиса...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при выключении HTTP сервера: %v", err)
	}
	log.Println("Сервис остановлен.")
}

// consumeKafka подписывается на Kafka и пишет в БД и кеш
func consumeKafka(cfg *Config, c *cache.Cache) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.KafkaTopic,
		GroupID: "order-service",
	})

	log.Printf("Подписка на Kafka (%s / %s)...", cfg.KafkaBroker, cfg.KafkaTopic)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Ошибка чтения из Kafka: %v", err)
			continue
		}

		var order db.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("Ошибка парсинга сообщения: %v", err)
			continue
		}

		// Сохраняем в БД
		if err := db.InsertOrder(&order); err != nil {
			log.Printf("Ошибка записи заказа в БД: %v", err)
			continue
		}

		// Обновляем кеш
		c.Set(order.OrderUID, order)

		log.Printf("Новый заказ сохранён: %s", order.OrderUID)
	}
}

// serveWeb — простая HTML страница
func serveWeb(w http.ResponseWriter, r *http.Request) {
	page := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Order Lookup</title>
	</head>
	<body>
		<h1>Поиск заказа</h1>
		<input type="text" id="orderId" placeholder="Введите OrderUID">
		<button onclick="fetchOrder()">Найти</button>
		<pre id="result"></pre>
		<script>
			function fetchOrder() {
				const id = document.getElementById("orderId").value;
				fetch("/order/" + id)
					.then(res => res.json())
					.then(data => {
						document.getElementById("result").textContent = JSON.stringify(data, null, 2);
					})
					.catch(err => {
						document.getElementById("result").textContent = "Ошибка: " + err;
					});
			}
		</script>
	</body>
	</html>`
	fmt.Fprint(w, page)
}
