CREATE TABLE IF NOT EXISTS deliveries (
    order_uid TEXT PRIMARY KEY,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

CREATE TABLE IF NOT EXISTS payments (
    transaction TEXT PRIMARY KEY,
    currency TEXT,
    provider TEXT,
    amount INT,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    date_created TIMESTAMP,
    FOREIGN KEY (order_uid) REFERENCES deliveries(order_uid),
    FOREIGN KEY (order_uid) REFERENCES payments(transaction)
);

CREATE TABLE IF NOT EXISTS items (
    chrt_id INT,
    track_number TEXT,
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id INT,
    brand TEXT,
    status INT,
    order_uid TEXT,
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);
