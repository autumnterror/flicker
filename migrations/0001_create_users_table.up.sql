CREATE TABLE users
(
    id    VARCHAR(50) PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(50) UNIQUE NOT NULL,
    about TEXT,
    password VARCHAR(100) NOT NULL,
    photo VARCHAR(100)
);