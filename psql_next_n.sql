INSERT INTO {{.Table}} (key, value)
VALUES ($1::text, $2::bigint)
ON CONFLICT (key)
DO UPDATE SET
   value = {{.Table}}.value + $2::bigint,
   updated_at = now()
   RETURNING value;
