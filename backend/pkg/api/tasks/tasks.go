package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	tasksdb "github.com/8run0/kllla2/backend/pkg/db/tasks"
	"github.com/8run0/kllla2/backend/pkg/svc/tasks"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func NewResource(db *tasksdb.DB) *Resource {
	return &Resource{
		Svc: &tasks.Service{
			DB: db,
		},
	}
}

type Resource struct {
	Svc *tasks.Service
}

const Path = "/tasks"

func (tr Resource) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", tr.List())
	r.Post("/", tr.Create())

	r.Route("/{id}", func(r chi.Router) {
		r.Use(tr.TaskCtx)
		r.Get("/", tr.Get())
		r.Put("/", tr.Update())
		r.Post("/", tr.Complete())
		r.Delete("/", tr.Delete())
	})
	return r
}

const TaskCtxIDKey = "id"

func (rs Resource) TaskCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := validateIDParam(idParam)
		if err != nil {
			http.Error(w, "invalid id param", http.StatusBadRequest)
			return
		}
		_, err = rs.Svc.GetByID(r.Context(), tasks.GetTaskRequest{
			ID: id,
		})
		if err != nil {
			http.Error(w, "task id not found", http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), TaskCtxIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTaskCtxID(ctx context.Context, w http.ResponseWriter, r *http.Request) uint64 {
	id, ok := ctx.Value(TaskCtxIDKey).(uint64)
	if !ok {
		http.Error(w, "task id not found", http.StatusNotFound)
	}
	return id
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Create implements taskRouter
func (rs Resource) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := &CreateTaskRequest{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			http.Error(w, "decode error", http.StatusInternalServerError)
		}
		id, err := rs.Svc.Create(ctx, tasks.CreateTaskRequest{
			Task: tasks.Task{
				Title:       req.Title,
				Description: req.Description,
			},
		})
		if err != nil {
			http.Error(w, "create error", http.StatusInternalServerError)
		}
		render.JSON(w, r, id)
	}
}

// GetByID implements taskRouter
func (rs Resource) Get() http.HandlerFunc {
	t, _ := template.ParseFiles("tasks.html")
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := getTaskCtxID(ctx, w, r)
		task, err := rs.Svc.GetByID(ctx, tasks.GetTaskRequest{
			ID: id,
		})
		if err != nil {
			http.Error(w, "get by id error", http.StatusInternalServerError)
			return
		}
		//render.JSON(w, r, task)
		t.Execute(w, task)
	}
}

func validateIDParam(idParam string) (id uint64, err error) {
	id, err = strconv.ParseUint(idParam, 10, 64)
	if id <= 0 || err != nil {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

// List implements taskRouter
func (rs Resource) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tasks, err := rs.Svc.List(ctx, tasks.ListTasksRequest{
			Limit: 10,
		})
		if err != nil {
			http.Error(w, "list error", http.StatusInternalServerError)
		}
		render.JSON(w, r, tasks)
	}
}

// Update implements taskResource
func (rs Resource) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := getTaskCtxID(ctx, w, r)
		req := &CreateTaskRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "decode error", http.StatusInternalServerError)
		}
		if err := rs.Svc.Update(ctx,
			tasks.UpdateTaskRequest{
				ID: id,
				Task: tasks.Task{
					Title:       req.Title,
					Description: req.Description,
				},
			}); err != nil {
			http.Error(w, "update error", http.StatusInternalServerError)
			return
		}
	}
}

// Delete implements taskRouter
func (rs Resource) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := getTaskCtxID(ctx, w, r)
		if err := rs.Svc.Delete(ctx, tasks.DeleteTaskRequest{
			ID: id,
		}); err != nil {
			http.Error(w, "compete error", http.StatusInternalServerError)
		}
	}
}

// Delete implements taskRouter
func (rs Resource) Complete() http.HandlerFunc {
	t, _ := template.ParseFiles("tasks.html")
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := getTaskCtxID(ctx, w, r)
		if err := rs.Svc.Complete(ctx, tasks.CompleteTaskRequest{
			ID: id,
		}); err != nil {
			http.Error(w, "complete error"+err.Error(), http.StatusInternalServerError)
		}
		task, err := rs.Svc.GetByID(ctx, tasks.GetTaskRequest{
			ID: id,
		})
		if err != nil {
			http.Error(w, "get updated model error"+err.Error(), http.StatusInternalServerError)
		}

		t.Execute(w, task)
	}
}
