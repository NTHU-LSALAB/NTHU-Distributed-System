CREATE TABLE IF NOT EXISTS comments(
    id serial PRIMARY KEY,
    video_id text NOT NULL,
    comment text NOT NULl
);