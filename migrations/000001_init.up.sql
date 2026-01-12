CREATE TABLE IF NOT EXISTS wishes (
    id SERIAL PRIMARY KEY,
    owner_email VARCHAR(100) NOT NULL,
    title VARCHAR(100) NOT NULL,
    description VARCHAR(300),
    is_bought BOOLEAN NOT NULL DEFAULT false,
    bought_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);