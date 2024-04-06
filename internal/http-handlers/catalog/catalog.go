package catalog

import (
	"catalog/internal/storage/entities"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	postgres "catalog/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

var (
	errPageInOutOfRange = errors.New("page in out of range")
)

type Request struct {
	entities.Car
	Page int `json:"page,omitempty"`
}

type Response struct {
	entities.CatalogPage
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.catalog.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		var err error

		if rawCarID := r.URL.Query().Get("carId"); rawCarID != "" {
			req.CarID, err = strconv.Atoi(rawCarID)
			if err != nil {
				log.Debug("failed to make int carId", err)
				w.WriteHeader(500)
				return
			}
		}
		req.Mark = r.URL.Query().Get("mark")
		req.Model = r.URL.Query().Get("model")
		if rawYear := r.URL.Query().Get("year"); rawYear != "" {
			req.Year, err = strconv.Atoi(rawYear)
			if err != nil {
				log.Debug("failed to make int year", err)
				w.WriteHeader(500)
				return
			}
		}
		req.Owner.Name = r.URL.Query().Get("name")
		req.Owner.Surname = r.URL.Query().Get("surname")
		req.Owner.Patronymic = r.URL.Query().Get("patronymic")
		if rawPage := r.URL.Query().Get("page"); rawPage != "" {
			req.Page, err = strconv.Atoi(rawPage)
			if err != nil {
				log.Debug("failed to make int page", err)
				w.WriteHeader(500)
				return
			}
		}

		// Get catalog on needed page with filter by c
		c := req.Car
		var cp entities.CatalogPage
		err = cp.GetCatalogPage(storage, &c, req.Page)
		// Case with page in out of range
		if errors.As(err, &errPageInOutOfRange) {
			log.Debug("failed to get catalog", err)
			w.WriteHeader(400)
			render.JSON(w, r, "Error: selected page in out of range")
			return
		}
		// Case with common error
		if err != nil {
			log.Debug("failed to get catalog", err)
			w.WriteHeader(500)
			return
		}

		log.Debug("catalog was successfully gotten on page " + strconv.Itoa(cp.Pagination.CurrentPage))

		rawResponse := Response{CatalogPage: cp}

		response, err := json.Marshal(rawResponse)
		if err != nil {
			log.Error("failed to code JSON response", err)
			w.WriteHeader(500)
			return
		}
		render.Data(w, r, response)

		log.Info("catalog response sent")
	}
}
