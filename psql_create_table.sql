CREATE TABLE IF NOT EXISTS {{.Table}} (
    key text primary key,
    value bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone
);
