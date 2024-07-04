-- name: CreateUserDb :one
INSERT INTO users (
        id,
        first_name,
        last_name,
        full_name,
        email,
        password,
        job_title,
        is_deleted,
        create_date,
        update_date
    )
VALUES (
        gen_random_uuid(),
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        false,
        now(),
        now()
    )
RETURNING id,
    first_name,
    last_name,
    full_name,
    email,
    job_title,
    is_deleted,
    create_date,
    update_date;
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;
-- name: GetUserIdByEmail :one
SELECT id
FROM users
WHERE email = $1;
-- name: GetUserByIdDb :one
SELECT id,
    first_name,
    last_name,
    full_name,
    email,
    password,
    job_title,
    is_deleted,
    create_date,
    update_date
FROM users
WHERE id = $1;
-- name: UserByIdDb :one
SELECT id,
    first_name,
    last_name,
    full_name,
    email,
    job_title,
    is_deleted,
    create_date,
    update_date
FROM users
WHERE id = $1;
-- name: DeleteUserDb :one
UPDATE users
SET is_deleted = true,
    update_date = now()
WHERE id = $1
RETURNING id,
    first_name,
    last_name,
    full_name,
    email,
    job_title,
    is_deleted,
    create_date,
    update_date;
-- name: EditPassword :exec
UPDATE users
SET password = $1,
    update_date = now()
WHERE id = $2;
-- name: GetUsersDb :many
SELECT id,
    first_name,
    last_name,
    full_name,
    email,
    job_title,
    is_deleted,
    create_date,
    update_date
FROM users;
-- name: UpdateUser :one
UPDATE users
SET first_name = $1,
    last_name = $2,
    full_name = $3,
    email = $4,
    job_title = $5,
    update_date = now()
WHERE id = $6
RETURNING id,
    first_name,
    last_name,
    full_name,
    email,
    job_title,
    is_deleted,
    create_date,
    update_date;