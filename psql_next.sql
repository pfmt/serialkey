INSERT INTO {{.Table}} (key, value)
VALUES ($1::text, $2::bigint)
ON CONFLICT (key)
DO UPDATE SET
   value = {{.Table}}.value + 1,
   updated_at = now()
   RETURNING value;
