CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    created_by TEXT NOT NULL,
    status TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    workflow_id TEXT,
    service_id INT,
    service_type TEXT,
    service_name TEXT,
    sub_status TEXT DEFAULT '',
    promotion_code TEXT DEFAULT '',
    fee_id TEXT DEFAULT '',
    has_insurance BOOLEAN DEFAULT FALSE,
    order_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_time TIMESTAMP WITH TIME ZONE,
    cancel_time TIMESTAMP WITH TIME ZONE,
    platform TEXT DEFAULT '',
    is_schedule BOOLEAN DEFAULT FALSE,
    now_order BOOLEAN DEFAULT FALSE,
    now_order_code TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_workflow_id ON orders(workflow_id);
