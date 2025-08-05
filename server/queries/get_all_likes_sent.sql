SELECT users.*
FROM decisions
INNER JOIN users ON decisions.to_user = users.id
WHERE decisions.from_user = $1 AND decisions.liked = true