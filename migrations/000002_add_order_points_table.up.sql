CREATE TABLE IF NOT EXISTS order_points (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    address TEXT NOT NULL,
    type VARCHAR(50) NOT NULL, -- e.g., 'pickup', 'dropoff'
    ordering INT NOT NULL DEFAULT 0, -- To maintain order of points: pickup -> dropoff
    CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE INDEX idx_order_points_order_id ON order_points(order_id);
