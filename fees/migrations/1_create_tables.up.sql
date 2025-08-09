-- Create the bills table
CREATE TABLE bills (
                       id TEXT PRIMARY KEY,
                       customer_id TEXT NOT NULL,
                       currency TEXT NOT NULL,
                       status TEXT NOT NULL,
                       total_amount BIGINT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL
);

-- Create the line_items table
CREATE TABLE line_items (
                            id SERIAL PRIMARY KEY,
                            bill_id TEXT NOT NULL REFERENCES bills(id),
                            description TEXT NOT NULL,
                            amount BIGINT NOT NULL,
                            timestamp TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_line_items_bill_id ON line_items (bill_id);