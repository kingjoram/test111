package delivery

import (
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/usecase"
)

type API struct {
	core usecase.ICore
	lg   *slog.Logger
	mx   *http.ServeMux
}

func GetApi(c *usecase.Core, l *slog.Logger) *API {
	api := &API{
		core: c,
		lg:   l.With("module", "api"),
	}
	mx := http.NewServeMux()
	mx.HandleFunc("/api/v1/csrf", api.GetCsrfToken)

	api.mx = mx

	return api
}

func (a *API) GetCsrfToken(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}

	csrfToken := r.Header.Get("x-csrf-token")

	found, err := a.core.CheckCsrfToken(r.Context(), csrfToken)
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}
	if csrfToken != "" && found {
		w.Header().Set("X-CSRF-Token", csrfToken)
		requests.SendResponse(w, response, a.lg)
		return
	}

	token, err := a.core.CreateCsrfToken(r.Context())
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	w.Header().Set("X-CSRF-Token", token)
	requests.SendResponse(w, response, a.lg)
}

func (a *API) ListenAndServe() {
	err := http.ListenAndServe(":8080", a.mx)
	if err != nil {
		a.lg.Error("ListenAndServe error", "err", err.Error())
	}
}
