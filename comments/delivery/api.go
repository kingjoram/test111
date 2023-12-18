package delivery

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/comments/usecase"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/metrics"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/middleware"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type API struct {
	core   usecase.ICore
	lg     *slog.Logger
	mx     *http.ServeMux
	mw     *middleware.ResponseMiddleware
	adress string
}

func GetApi(c *usecase.Core, l *slog.Logger, cfg *configs.CommentCfg) *API {
	resp := &requests.Response{
		Status: http.StatusOK,
		Body:   nil,
	}

	middleware := &middleware.ResponseMiddleware{
		Response: resp,
		Metrix:   metrics.GetMetrics(),
	}

	api := &API{
		core:   c,
		lg:     l.With("module", "api"),
		mx:     http.NewServeMux(),
		mw:     middleware,
		adress: cfg.ServerAdress,
	}

	api.mx.Handle("/metrics", promhttp.Handler())
	api.mx.Handle("/api/v1/comment", api.mw.GetResponse(http.HandlerFunc(api.Comment), l))
	api.mx.Handle("/api/v1/comment/add", api.mw.GetResponse(http.HandlerFunc(api.AddComment), l))

	return api
}

func (a *API) ListenAndServe() {
	err := http.ListenAndServe(a.adress, a.mx)
	if err != nil {
		a.lg.Error("listen and serve error", "err", err.Error())
	}
}

func (a *API) Comment(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		a.mw.Response.Status = http.StatusMethodNotAllowed
		return
	}

	filmId, err := strconv.ParseUint(r.URL.Query().Get("film_id"), 10, 64)
	if err != nil {
		a.mw.Response.Status = http.StatusBadRequest
		return
	}
	page, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.ParseUint(r.URL.Query().Get("per_page"), 10, 64)
	if err != nil {
		pageSize = 10
	}

	comments, err := a.core.GetFilmComments(filmId, (page-1)*pageSize, pageSize)
	if err != nil {
		a.lg.Error("Comment", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	commentsResponse := requests.CommentResponse{Comments: comments}

	a.mw.Response.Body = commentsResponse
}

func (a *API) AddComment(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		a.mw.Response.Status = http.StatusMethodNotAllowed
		return
	}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		a.mw.Response.Status = http.StatusUnauthorized
		return
	}
	if err != nil {
		a.lg.Error("Add comment error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	userId, err := a.core.GetUserId(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("Add comment error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	var commentRequest requests.CommentRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	if err = json.Unmarshal(body, &commentRequest); err != nil {
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	found, err := a.core.AddComment(commentRequest.FilmId, userId, commentRequest.Rating, commentRequest.Text)
	if err != nil {
		a.lg.Error("Add Comment error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
	}
	if found {
		a.mw.Response.Status = http.StatusNotAcceptable
		return
	}
}
