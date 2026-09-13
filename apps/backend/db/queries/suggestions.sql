-- name: CreateProductSuggestion :one
WITH inserted AS (
    INSERT INTO product_suggestions (account_id, body)
    VALUES (sqlc.arg(account_id)::uuid, sqlc.arg(body))
    RETURNING id, account_id, body, created_at
)
SELECT
    inserted.id::text AS id,
    accounts.username,
    inserted.body,
    inserted.created_at
FROM inserted
JOIN accounts ON accounts.id = inserted.account_id;
