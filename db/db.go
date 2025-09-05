package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB — инициализация подключения к PostgreSQL
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("PostgreSQL не отвечает: %w", err)
	}

	log.Println("Подключение к PostgreSQL успешно!")
	return nil
}

// GetAllOrders — достаём все заказы с джойном доставок и платежей
func GetAllOrders() ([]Order, error) {
	rows, err := DB.Query(`
	SELECT o.order_uid, o.track_number, o.entry, o.locale, o.customer_id,
	       o.delivery_service, o.date_created,
	       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	       p.transaction, p.currency, p.provider, p.amount, p.payment_dt, p.bank,
	       p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
	JOIN deliveries d ON o.order_uid = d.order_uid
	JOIN payments   p ON o.order_uid = p.order_uid
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		var d Delivery
		var p Payment

		err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale,
			&o.CustomerID, &o.DeliveryService, &o.DateCreated,
			&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.Transaction, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDT,
			&p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
		)
		if err != nil {
			return nil, err
		}

		// items для заказа
		items, err := GetItemsByOrder(o.OrderUID)
		if err != nil {
			return nil, err
		}

		o.Delivery = d
		o.Payment = p
		o.Items = items

		orders = append(orders, o)
	}
	return orders, nil
}

// GetItemsByOrder — достаём список товаров по заказу
func GetItemsByOrder(orderUID string) ([]Item, error) {
	rows, err := DB.Query(`
	SELECT chrt_id, track_number, price, rid, name, sale, size,
	       total_price, nm_id, brand, status
	FROM items
	WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		err := rows.Scan(
			&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name,
			&it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

// GetOrderByUID — поиск заказа по UID
func GetOrderByUID(orderUID string) (*Order, error) {
	row := DB.QueryRow(`
	SELECT o.order_uid, o.track_number, o.entry, o.locale, o.customer_id,
	       o.delivery_service, o.date_created,
	       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	       p.transaction, p.currency, p.provider, p.amount, p.payment_dt, p.bank,
	       p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
	JOIN deliveries d ON o.order_uid = d.order_uid
	JOIN payments   p ON o.order_uid = p.order_uid
	WHERE o.order_uid = $1
	`, orderUID)

	var o Order
	var d Delivery
	var p Payment

	err := row.Scan(
		&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale,
		&o.CustomerID, &o.DeliveryService, &o.DateCreated,
		&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
		&p.Transaction, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDT,
		&p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	items, err := GetItemsByOrder(o.OrderUID)
	if err != nil {
		return nil, err
	}

	o.Delivery = d
	o.Payment = p
	o.Items = items

	return &o, nil
}

// InsertOrder — сохраняем заказ и его зависимости (delivery, payment, items)
func InsertOrder(order *Order) error {
	// orders
	_, err := DB.Exec(`
		INSERT INTO orders (order_uid, track_number, entry, locale, customer_id,
		                    delivery_service, date_created)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.CustomerID, order.DeliveryService, order.DateCreated)
	if err != nil {
		return fmt.Errorf("ошибка вставки заказа: %w", err)
	}

	// delivery
	_, err = DB.Exec(`
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("ошибка вставки доставки: %w", err)
	}

	// payment
	_, err = DB.Exec(`
		INSERT INTO payments (transaction, currency, provider, amount, payment_dt,
		                      bank, delivery_cost, goods_total, custom_fee, order_uid)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, order.Payment.Transaction, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
		order.OrderUID)
	if err != nil {
		return fmt.Errorf("ошибка вставки платежа: %w", err)
	}

	// items
	for _, it := range order.Items {
		_, err = DB.Exec(`
			INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size,
			                   total_price, nm_id, brand, status, order_uid)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name, it.Sale,
			it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status, order.OrderUID)
		if err != nil {
			return fmt.Errorf("ошибка вставки товара: %w", err)
		}
	}

	return nil
}
