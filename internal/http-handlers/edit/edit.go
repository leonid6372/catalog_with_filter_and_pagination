package edit

import (
	"catalog/internal/lib/logger/sl"
	"catalog/internal/storage/entities"
	"errors"
	"io"
	"log/slog"
	"net/http"

	postgres "catalog/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	CarID  int             `json:"carId",omitempty validate:"required""`
	RegNum string          `json:"regNum",omitempty`
	Mark   string          `json:"mark",omitempty`
	Model  string          `json:"model",omitempty`
	Year   int             `json:"year",omitempty`
	Owner  entities.Person `json:"owner",omitempty`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.edit.New"

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

		c := entities.Car{
			CarID:  req.CarID,
			RegNum: req.RegNum,
			Mark:   req.Mark,
			Model:  req.Model,
			Year:   req.Year,
			Owner:  req.Owner,
		}
		if err := c.Edit(storage); err != nil {
			w.WriteHeader(500)
			log.Debug("failed to edit car", sl.Err(err))
			return
		}

		log.Debug("car was successfully edited")
	}
}
