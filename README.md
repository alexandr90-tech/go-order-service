# Go Order Service

Сервис для приёма и обработки заказов. Архитектура построена на Go, Kafka и PostgreSQL.  
Заказы генерируются через Kafka Producer, обрабатываются Consumer, сохраняются в БД и кэшируются в памяти.  
Доступ к данным реализован через REST API и простую HTML-страницу.

---

## 🚀 Стек технологий
- **Go** (net/http, gorilla/mux, kafka-go, sarama)
- **PostgreSQL** (хранение заказов)
- **Kafka** (очередь сообщений)
- **Docker Compose** (поднимает окружение: PostgreSQL, Kafka, Zookeeper)

---

## ⚙️ Запуск проекта

### 1. Поднять инфраструктуру
```bash
docker compose up -d
```
Будут запущены контейнеры:
PostgreSQL (`localhost:5432`)
Zookeeper (`localhost:2181`)
Kafka (`localhost:9092`)

### 2. Запустить сервис
```bash
go run .
```
В логах будет видно:
подключение к PostgreSQL
применение миграций
подписка на Kafka
запуск HTTP сервера (:8080)

### 3. Отправить заказ в Kafka
в новом терминале
```bash
go run ./cmd/producer
```
Лог сервиса:
```sql
Новый заказ сохранён: order-123456
```
### 4. Проверить заказ
Через API:
```http
GET http://localhost:8080/order/order-123456
```
Ответ: JSON с заказом.

Через веб-страницу:
Открыть в браузере → http://localhost:8080/
Ввести OrderUID → получить JSON.

### 📂 Структура проекта
```bash
go-order-service/
│── cmd/
│   └── producer/        # Kafka Producer для генерации заказов
│── db/
│   ├── migrate.go       # SQL миграции (создание таблиц)
│   ├── db.go            # Логика работы с PostgreSQL
│   └── models.go        # Модели заказов
│── cache/
│   └── cache.go         # In-memory кэш заказов
│── docker-compose.yml   # Инфраструктура (Postgres, Kafka, Zookeeper)
│── main.go              # Основной сервис (consumer + API + cache)

```
### ✅Функционал
Приём заказов через Kafka

Запись заказов в PostgreSQL

In-memory кэш для ускорения доступа

REST API для получения заказа

HTML-страница для поиска заказа
