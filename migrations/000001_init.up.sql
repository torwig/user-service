CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    first_name varchar(255) NOT NULL,
    last_name varchar(255) NOT NULL,
    phone_number varchar(30) NOT NULL,
    address varchar(255) NOT NULL,
    deleted boolean NOT NULL DEFAULT FALSE,
    created_at timestamp NOT NULL DEFAULT NOW(),
    deleted_at timestamp DEFAULT NULL
);