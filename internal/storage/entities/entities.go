package entities

import (
	postgres "catalog/internal/storage"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
)

const (
	qrNewCar       = `INSERT INTO car(reg_num, mark, model, year, owner) VALUES ($1, $2, $3, $4, $5);`
	qrDelete       = `DELETE FROM car WHERE car_id = $1;`
	qrGetCarsCount = `SELECT count("car_id") FROM car;`
	qrGetPersonID  = `SELECT person_id FROM person WHERE "name" = $1 AND surname = $2 AND patronymic = $3;`
	qrGetPerson    = `SELECT "name", surname, patronymic FROM person WHERE person_id = $1;`
	qrNewPerson    = `INSERT INTO person("name", surname, patronymic) VALUES ($1, $2, $3)
				   	  ON CONFLICT ("name", surname, patronymic) DO NOTHING;`
)

type Person struct {
	PersonID   int    `json:"personId,omitempty"`
	Name       string `json:"name,omitempty"`
	Surname    string `json:"surname,omitempty"`
	Patronymic string `json:"patronymic,omitempty"`
}

type Car struct {
	CarID  int    `json:"carId,omitempty" validate:"required"`
	RegNum string `json:"regNum,omitempty"`
	Mark   string `json:"mark,omitempty"`
	Model  string `json:"model,omitempty"`
	Year   int    `json:"year,omitempty"`
	Owner  Person `json:"owner,omitempty"`
}

func (c *Car) Delete(storage *postgres.Storage, carID int) error {
	const op = "storage.entities.Delete"

	_, err := storage.DB.Exec(qrDelete, carID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Car) Edit(storage *postgres.Storage) error {
	const op = "storage.entities.Edit"

	qrEdit := "UPDATE car SET "
	qrParameters := []interface{}{}

	var emptyCar Car
	i := 0
	if c.RegNum != emptyCar.RegNum {
		i++
		qrEdit += " reg_num = $" + strconv.Itoa(i) + " "
		qrParameters = append(qrParameters, c.RegNum)
	}
	if c.Mark != emptyCar.Mark {
		if i != 0 {
			qrEdit += ", "
		}
		i++
		qrEdit += " mark = $" + strconv.Itoa(i) + " "
		qrParameters = append(qrParameters, c.Mark)
	}
	if c.Model != emptyCar.Model {
		if i != 0 {
			qrEdit += ", "
		}
		i++
		qrEdit += " model = $" + strconv.Itoa(i) + " "
		qrParameters = append(qrParameters, c.Model)
	}
	if c.Year != emptyCar.Year {
		if i != 0 {
			qrEdit += ", "
		}
		i++
		qrEdit += ` "year" = $` + strconv.Itoa(i) + " "
		qrParameters = append(qrParameters, c.Year)
	}
	if c.Owner != emptyCar.Owner {
		// Validate request JSON
		if err := validator.New().Struct(c.Owner); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		_, err := storage.DB.Exec(qrNewPerson, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		var personID int
		err = storage.DB.QueryRow(qrGetPersonID, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic).Scan(&personID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if i != 0 {
			qrEdit += ", "
		}
		qrEdit += ` "owner" = ` + strconv.Itoa(personID) + " "
	}
	qrEdit += ` WHERE car_id = ` + strconv.Itoa(c.CarID) + ";"

	_, err := storage.DB.Exec(qrEdit, qrParameters...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type Cars []Car

type CatalogPage struct {
	Cars
	Pagination Pagination
}

func (cp *CatalogPage) GetCatalogPage(storage *postgres.Storage, c *Car, page int) error {
	const op = "storage.entities.GetCatalogPage"

	qrGetCars := "SELECT * FROM car"
	qrParameters := []interface{}{}
	var recordsCount int
	i := 0 // i - query parameters counter

	var emptyCar Car
	if *c != emptyCar {
		qrGetCars += " WHERE "
		if c.CarID != emptyCar.CarID {
			qrGetCars += " car_id = $" + strconv.Itoa(i) + " "
			qrParameters = append(qrParameters, c.CarID)
		}
		if c.RegNum != emptyCar.RegNum {
			if i != 0 {
				qrGetCars += " AND "
			}
			i++
			qrGetCars += " reg_num = $" + strconv.Itoa(i) + " "
			qrParameters = append(qrParameters, c.RegNum)
		}
		if c.Mark != emptyCar.Mark {
			if i != 0 {
				qrGetCars += " AND "
			}
			i++
			qrGetCars += " mark = $" + strconv.Itoa(i) + " "
			qrParameters = append(qrParameters, c.Mark)
		}
		if c.Model != emptyCar.Model {
			if i != 0 {
				qrGetCars += " AND "
			}
			i++
			qrGetCars += " model = $" + strconv.Itoa(i) + " "
			qrParameters = append(qrParameters, c.Model)
		}
		if c.Year != emptyCar.Year {
			if i != 0 {
				qrGetCars += " AND "
			}
			i++
			qrGetCars += ` "year" = $` + strconv.Itoa(i) + " "
			qrParameters = append(qrParameters, c.Year)
		}
		if c.Owner != emptyCar.Owner {
			// Validate request JSON
			if err := validator.New().Struct(c.Owner); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			_, err := storage.DB.Exec(qrNewPerson, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			var personID int
			err = storage.DB.QueryRow(qrGetPersonID, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic).Scan(&personID)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if i != 0 {
				qrGetCars += " AND "
			}
			qrGetCars += ` "owner" = ` + strconv.Itoa(personID) + " "
		}

		// Get filtered records count
		if err := storage.DB.QueryRow(("SELECT count(car_id) FROM car WHERE " + qrGetCars[24:] + ";"), qrParameters...).Scan(&recordsCount); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	// Count all records it there was not filter
	if i == 0 {
		// Get all records count
		if err := storage.DB.QueryRow(`SELECT count(car_id) FROM car;`).Scan(&recordsCount); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	// Make pagination
	if page < 0 {
		return fmt.Errorf("%s: page in out of range", op)
	}
	if page == 0 {
		page = 1
	}
	limit := 2
	offset := limit * (page - 1)
	if err := cp.Pagination.NewPagination(recordsCount, limit, page); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	i++
	qrGetCars += ` order by car_id desc limit $` + strconv.Itoa(i) + " "
	qrParameters = append(qrParameters, limit)
	i++
	qrGetCars += " offset $" + strconv.Itoa(i)
	qrParameters = append(qrParameters, offset)
	qrGetCars += ";"

	qrResult, err := storage.DB.Query(qrGetCars, qrParameters...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	for qrResult.Next() {
		var c Car
		var ownerID int
		if err := qrResult.Scan(&c.CarID, &c.RegNum, &c.Mark, &c.Model, &c.Year, &ownerID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if err := storage.DB.QueryRow(qrGetPerson, ownerID).Scan(&c.Owner.Name, &c.Owner.Surname, &c.Owner.Patronymic); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		cp.Cars = append(cp.Cars, c)
	}

	return nil
}

func (c *Car) New(storage *postgres.Storage) error {
	const op = "storage.entities.New"

	_, err := storage.DB.Exec(qrNewPerson, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var personID int
	err = storage.DB.QueryRow(qrGetPersonID, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic).Scan(&personID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = storage.DB.Exec(qrNewCar, c.RegNum, c.Mark, c.Model, c.Year, personID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type Pagination struct {
	Next          int
	Previous      int
	RecordPerPage int
	CurrentPage   int
	TotalPage     int
}

// Generated Pagination Meta data
func (p *Pagination) NewPagination(recordsCount, limit, page int) error {
	const op = "storage.entities.NewPagination"

	total := (recordsCount / limit)

	// Calculator Total Page
	remainder := (recordsCount % limit)
	if remainder == 0 {
		p.TotalPage = total
	} else {
		p.TotalPage = total + 1
	}

	if page > p.TotalPage {
		return fmt.Errorf("%s: page in out of range", op)
	}

	// Set current/record per page meta data
	p.CurrentPage = page
	p.RecordPerPage = limit

	// Calculator the Next/Previous Page
	if page <= 0 {
		p.Next = page + 1
	} else if page < p.TotalPage {
		p.Previous = page - 1
		p.Next = page + 1
	} else if page == p.TotalPage {
		p.Previous = page - 1
		p.Next = 0
	}

	return nil
}
