CREATE TABLE users (
    id          UUID PRIMARY KEY    DEFAULT gen_random_uuid(),
    created_at  TIMESTAMP           DEFAULT NOW()
);

CREATE TABLE decisions (
    id          SERIAL PRIMARY KEY,
    created_at  TIMESTAMP           DEFAULT NOW(),
    from_user   UUID NOT NULL       REFERENCES users(id),
    to_user     UUID NOT NULL       REFERENCES users(id),
    liked       BOOLEAN NOT NULL,
    UNIQUE(from_user, to_user)
);
