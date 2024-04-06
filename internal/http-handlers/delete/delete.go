package delete

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"catalog/internal/lib/logger/sl"
	postgres "catalog/internal/storage"
	"catalog/internal/storage/entities"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	CarID int `json:"carId" validate:"required"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"

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

		log.Info("request body decoded", slog.Any("request", req))

		// Validate request JSON
		if err := validator.New().Struct(req); err != nil {
			w.WriteHeader(400)
			log.Error("invalid request", sl.Err(err))
			return
		}

		var c *entities.Car
		if err := c.Delete(storage, req.CarID); err != nil {
			w.WriteHeader(500)
			log.Debug("failed to delete car", sl.Err(err))
			return
		}

		log.Debug("car was successfully deleted")
	}
}
