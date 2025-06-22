package create

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	resp "github.com/Noviiich/io-bound-task/internal/lib/api/response"
	"github.com/Noviiich/io-bound-task/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

type Response struct {
	ID string `json:"id"`
	resp.Response
}

type TaskRegisterService interface {
	RegisterTask(name string) (string, error)
}

func New(log *slog.Logger, taskRegister TaskRegisterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.tasks.create.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		id, err := taskRegister.RegisterTask(req.Name)
		if err != nil {
			log.Error("failed to create task", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to create task"))

			return
		}

		log.Info("task created successfully", slog.String("task_id", id))
		responseOK(w, r, id)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id string) {
	render.JSON(w, r, Response{
		Response: resp.Success(),
		ID:       id,
	})
}
