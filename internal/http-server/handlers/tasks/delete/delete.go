package delete

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Noviiich/io-bound-task/internal/lib/api/response"
	"github.com/Noviiich/io-bound-task/internal/lib/logger/sl"
	"github.com/Noviiich/io-bound-task/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type TaskDeleter interface {
	DeleteTask(id string) error
}

type Response struct {
	resp.Response
	ID string `json:"id"`
}

func New(log *slog.Logger, taskGetter TaskDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id := chi.URLParam(r, "id")
		log.Info("request received", slog.String("id", id))
		if id == "" {
			log.Info("id is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		err := taskGetter.DeleteTask(id)
		if errors.Is(err, storage.ErrTaskNotFound) {
			log.Info("task not found", "id", id)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to delete task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("delete task", slog.String("task", id))

		responseDeleted(w, r, id)
	}
}

func responseDeleted(w http.ResponseWriter, r *http.Request, id string) {
	render.JSON(w, r, Response{
		Response: resp.Deleted(),
		ID:       id,
	})
}
