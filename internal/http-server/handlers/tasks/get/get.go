package get

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Noviiich/io-bound-task/internal/domain/models"
	resp "github.com/Noviiich/io-bound-task/internal/lib/api/response"
	"github.com/Noviiich/io-bound-task/internal/lib/logger/sl"
	"github.com/Noviiich/io-bound-task/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type TaskGetter interface {
	GetTask(id string) (*models.Task, error)
}

type Response struct {
	models.TaskResponse
}

func New(log *slog.Logger, taskGetter TaskGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.get.New"

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

		task, err := taskGetter.GetTask(id)
		if errors.Is(err, storage.ErrTaskNotFound) {
			log.Info("task not found", "id", id)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("got task", slog.String("task", task.ID))

		responseStatus(w, r, *task.ToResponse())
	}
}

func responseStatus(w http.ResponseWriter, r *http.Request, taskResp models.TaskResponse) {
	render.JSON(w, r, Response{
		TaskResponse: taskResp,
	})
}
