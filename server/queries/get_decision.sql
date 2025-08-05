SELECT *
FROM decisions
WHERE decisions.from_user = $1 AND decisions.to_user = $2 AND decisions.liked = true