INSERT INTO decisions (from_user, to_user, liked)
VALUES ($1, $2, $3)
ON CONFLICT (from_user, to_user)
DO UPDATE SET liked = $3
RETURNING id, created_at, from_user, to_user, liked;