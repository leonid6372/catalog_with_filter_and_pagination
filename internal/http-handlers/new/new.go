package new

import (
	"catalog/internal/lib/logger/sl"
	"catalog/internal/storage/entities"
	"errors"
	"io"
	"log/slog"
	"net/http"

	postgres "catalog/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	RegNum string `json:"regNum" validate:"required"`
}

type PersonResponse struct {
	Name       string `json:"name,omitempty" validate:"required"`
	Surname    string `json:"surname,omitempty" validate:"required"`
	Patronymic string `json:"patronymic,omitempty"`
}

type CarResponse struct {
	RegNum string         `json:"regNum,omitempty" validate:"required"`
	Mark   string         `json:"mark,omitempty" validate:"required"`
	Model  string         `json:"model,omitempty" validate:"required"`
	Year   int            `json:"year,omitempty"`
	Owner  PersonResponse `json:"owner,omitempty" validate:"required"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.new.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Decode request JSON
		err := render.DecodeJSON(r.Body, &req)
		// Case with empty request
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			w.WriteHeader(400)
			return
		}
		// Case with common errors
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			w.WriteHeader(400)
			return
		}

		// Validate request JSON
		if err := validator.New().Struct(req); err != nil {
			w.WriteHeader(400)
			log.Error("invalid request", sl.Err(err))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// GET request to archive. Response about car with regNum
		resp, err := http.Get("http://localhost:8080/car_information?regNum=" + req.RegNum)
		if err != nil {
			log.Error("failed to make GET request to archive", sl.Err(err))
		}

		var cr CarResponse
		// Decode request JSON
		err = render.DecodeJSON(resp.Body, &cr)
		// Case with empty request
		if errors.Is(err, io.EOF) {
			log.Error("archive response body is empty")
			w.WriteHeader(400)
			return
		}

		log.Info("response body decoded", slog.Any("request", req))

		// Validate response JSON
		if err := validator.New().Struct(cr); err != nil {
			w.WriteHeader(400)
			log.Error("invalid request", sl.Err(err))
			return
		}

		o := entities.Person{
			Name:       cr.Owner.Name,
			Surname:    cr.Owner.Surname,
			Patronymic: cr.Owner.Patronymic,
		}

		c := entities.Car{
			RegNum: cr.RegNum,
			Mark:   cr.Mark,
			Model:  cr.Model,
			Year:   cr.Year,
			Owner:  o,
		}

		if err = c.New(storage); err != nil {
			w.WriteHeader(500)
			log.Debug("failed to add new car in catalog", sl.Err(err))
			return
		}
	}
}
