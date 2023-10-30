package http

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/torwig/user-service/entities"
	"github.com/torwig/user-service/ports/http/requests"
	"github.com/torwig/user-service/ports/http/responses"
	"go.uber.org/zap"
)

var (
	errEmptyParameter      = errors.New("parameter wasn't specified")
	errParameterNotInteger = errors.New("parameter isn't an integer value")
)

//go:embed docs
var docsFS embed.FS

type UserService interface {
	CreateUser(ctx context.Context, params entities.CreateUserParams) (entities.User, error)
	GetUser(ctx context.Context, id int64) (entities.User, error)
	UpdateUser(ctx context.Context, id int64, params entities.UpdateUserParams) (entities.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type UserAuthenticator interface {
	ParseAccessToken(t string) (*entities.AuthenticatedUser, error)
}

type Handler struct {
	svc  UserService
	auth UserAuthenticator
	log  *zap.SugaredLogger
}

func NewHandler(userSvc UserService, userAuth UserAuthenticator, log *zap.SugaredLogger) *Handler {
	return &Handler{svc: userSvc, auth: userAuth, log: log}
}

func (h *Handler) Router() http.Handler {
	docsAsRootFS, err := fs.Sub(docsFS, "docs")
	if err != nil {
		panic(fmt.Sprintf("failed to setup embed FS: %s", err))
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(BearerTokenAuthentication(h.auth))

	r.Handle("/docs/*", http.FileServer(http.FS(docsAsRootFS)))

	r.Get("/health", h.healthcheck)

	r.Route("/api/v1/users", func(r chi.Router) {
		r.Post("/", h.createUser)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getUser)
			r.Patch("/", h.updateUser)
			r.Delete("/", h.deleteUser)
		})
	})

	return r
}

func (h *Handler) healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	au, err := AuthenticatedUserFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !au.CanCreate() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	req, err := requests.NewCreateUser(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createdUser, err := h.svc.CreateUser(r.Context(), req.ToCreateUserParams())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responses.SendJSON(w, http.StatusCreated, responses.UserFromEntity(createdUser))
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := userIdentifierFromRequestURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	au, err := AuthenticatedUserFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !au.CanViewUser(id) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, entities.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	responses.SendJSON(w, http.StatusOK, responses.UserFromEntity(user))
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := userIdentifierFromRequestURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	au, err := AuthenticatedUserFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !au.CanUpdateUser(id) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	req, err := requests.NewUpdateUser(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedUser, err := h.svc.UpdateUser(r.Context(), id, req.ToUpdateUserParams())
	if err != nil {
		if errors.Is(err, entities.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	responses.SendJSON(w, http.StatusOK, responses.UserFromEntity(updatedUser))
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := userIdentifierFromRequestURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	au, err := AuthenticatedUserFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !au.CanDelete(id) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, entities.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userIdentifierFromRequestURL(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return 0, errEmptyParameter
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errParameterNotInteger
	}

	return id, nil
}
