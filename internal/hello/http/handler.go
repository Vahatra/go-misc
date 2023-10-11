package http

import (
	"context"
	"log/slog"
	"net/http"

	ihttp "go-misc/internal/http"

	"github.com/gorilla/mux"
)

type HelloService interface {
	Say(ctx context.Context, address string) (string, error)
}

type accountHandler struct {
	l *slog.Logger
	s HelloService
}

func NewHandler(l *slog.Logger, s HelloService) *accountHandler {
	return &accountHandler{l: l, s: s}
}

func (h *accountHandler) Register(r *mux.Router) {
	r.Path("/say/{id}").HandlerFunc(h.say).Methods("GET")
}

func (h *accountHandler) say(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	msg, err := h.s.Say(r.Context(), vars["id"])
	if err != nil {
		ihttp.EncodeError(r.Context(), w, err)
		return
	}
	resp := struct {
		Msg string `json:"message"`
	}{
		Msg: msg,
	}
	err = ihttp.EncodeResponse(w, resp)
	if err != nil {
		ihttp.EncodeError(r.Context(), w, err)
		return
	}
}
