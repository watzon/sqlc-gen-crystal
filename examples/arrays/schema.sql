CREATE TABLE authors (
  id         SERIAL PRIMARY KEY,
  name       text   NOT NULL,
  bio        text   NOT NULL,
  birth_year int    NOT NULL
);

CREATE TABLE books (
  id          SERIAL PRIMARY KEY,
  author_id   int    NOT NULL REFERENCES authors(id),
  title       text   NOT NULL,
  isbn        text   NOT NULL,
  published   date   NOT NULL,
  tags        text[] NOT NULL DEFAULT '{}'
);