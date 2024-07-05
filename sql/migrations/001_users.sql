-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NULL,
    full_name VARCHAR(255) NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    job_title VARCHAR(255) NULL,
    is_deleted bool NULL,
    create_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_DATE,
    update_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_DATE
);
-- +goose Down
DROP TABLE users;