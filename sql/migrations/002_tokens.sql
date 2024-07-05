-- +goose Up
CREATE TABLE tokens (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id UUID NOT NULL,
    token VARCHAR(600) NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose Down
DROP TABLE tokens;