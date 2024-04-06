CREATE TABLE IF NOT EXISTS person(
	person_id SERIAL PRIMARY KEY,
	"name" TEXT NOT NULL,
	surname	TEXT NOT NULL,
	patronymic	TEXT NOT NULL
);

ALTER TABLE person ADD CONSTRAINT fullname_constraint UNIQUE ("name", surname, patronymic);

CREATE TABLE IF NOT EXISTS car(
	car_id SERIAL PRIMARY KEY,
	reg_num TEXT NOT NULL,
	mark TEXT NOT NULL,
	model TEXT NOT NULL,
	"year" INT,
	"owner" INT NOT NULL REFERENCES person(person_id) ON DELETE CASCADE
);