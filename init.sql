CREATE TABLE IF NOT EXISTS payments (
    user_id VARCHAR(255) NOT NULL,
    package_id VARCHAR(255) NOT NULL,
    cost DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    PRIMARY KEY (user_id, package_id)
);
