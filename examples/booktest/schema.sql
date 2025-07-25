CREATE TABLE authors (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  bio        TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE books (
  id          SERIAL PRIMARY KEY,
  author_id   INTEGER NOT NULL REFERENCES authors(id),
  title       TEXT NOT NULL,
  description TEXT,
  price       DECIMAL(10,2) NOT NULL,
  isbn        VARCHAR(13),
  published   DATE,
  created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE reviews (
  id         SERIAL PRIMARY KEY,
  book_id    INTEGER NOT NULL REFERENCES books(id),
  reviewer   TEXT NOT NULL,
  rating     INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
  comment    TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);