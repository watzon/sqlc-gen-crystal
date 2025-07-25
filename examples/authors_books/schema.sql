CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  bio  text
);

CREATE TABLE books (
  id        BIGSERIAL PRIMARY KEY,
  author_id BIGINT    NOT NULL REFERENCES authors(id),
  title     text      NOT NULL,
  isbn      text      UNIQUE,
  published date
);