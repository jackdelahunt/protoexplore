SELECT COUNT(*)
FROM decisions
WHERE decisions.to_user = $1 AND decisions.liked = true