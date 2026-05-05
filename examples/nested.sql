CREATE TABLE orders (
  id BIGINT PRIMARY KEY,
  customer_email TEXT,
  customer_name TEXT
);

CREATE TABLE items (
  id BIGINT PRIMARY KEY,
  order_id BIGINT,
  product TEXT,
  qty DOUBLE PRECISION,
  FOREIGN KEY (order_id) REFERENCES orders(id)
);
