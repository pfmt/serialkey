WITH current AS (
     SELECT value FROM {{.Table}} WHERE key = $1::text
),   difference AS (
     SELECT ($2::bigint - coalesce(value, 0)) AS value FROM current
),   increment AS (
     SELECT (CASE WHEN value > 0 THEN value ELSE 0 END) + 1 AS value FROM difference
) INSERT INTO {{.Table}} (key, value)
  VALUES ($1::text, $2::bigint)
  ON CONFLICT (key)
  DO UPDATE SET
     value = {{.Table}}.value + (SELECT * FROM increment),
     updated_at = now()
     RETURNING value;
