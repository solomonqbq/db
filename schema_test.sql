CREATE TABLE users(
	id INTEGER NOT NULL PRIMARY KEY,
	name STRING UNIQUE,
	age INTEGER)
;

INSERT INTO users(id, name, age) VALUES
	(1, 'bob', 32),
	(2, 'mike', 25),
	(3, 'john', 55)
;
