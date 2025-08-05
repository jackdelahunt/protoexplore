SELECT users.*
FROM decisions
INNER JOIN users ON decisions.from_user = users.id
WHERE decisions.to_user = $1 AND decisions.liked = true AND users.id < $2
ORDER BY id DESC
LIMIT $3
