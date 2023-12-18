package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/usecase"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/metrics"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
)

type ResponseMiddleware struct {
	Response *requests.Response
	Metrix   *metrics.Metrics
}

type contextKey string

const UserIDKey contextKey = "userId"

func (mw *ResponseMiddleware) setResponseAndMetrics(w http.ResponseWriter, r *http.Request, status int, start time.Time, lg *slog.Logger) {
	mw.Response.Status = status
	end := time.Since(start)
	mw.Metrix.Time.WithLabelValues(strconv.Itoa(mw.Response.Status), r.URL.Path).Observe(end.Seconds())
	mw.Metrix.Hits.WithLabelValues(strconv.Itoa(mw.Response.Status), r.URL.Path).Inc()
	requests.SendResponse(w, *mw.Response, lg)
}

func (mw *ResponseMiddleware) AuthCheck(next http.Handler, core *usecase.Core, lg *slog.Logger) http.Handler {
	start := time.Now()
	mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session_id")
		if errors.Is(err, http.ErrNoCookie) {
			mw.setResponseAndMetrics(w, r, http.StatusUnauthorized, start, lg)
			return
		}

		userId, err := core.GetUserId(r.Context(), session.Value)
		if err != nil {
			lg.Error("auth check error", "err", err.Error())
			mw.setResponseAndMetrics(w, r, http.StatusInternalServerError, start, lg)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), UserIDKey, userId))

		next.ServeHTTP(w, r)
		requests.SendResponse(w, *mw.Response, lg)
		end := time.Since(start)
		mw.Metrix.Time.WithLabelValues(strconv.Itoa(mw.Response.Status), r.URL.Path).Observe(end.Seconds())
		mw.Metrix.Hits.WithLabelValues(strconv.Itoa(mw.Response.Status), r.URL.Path).Inc()
	})
}

func (mw *ResponseMiddleware) GetResponse(next http.Handler, lg *slog.Logger) http.Handler {
	start := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		requests.SendResponse(w, *mw.Response, lg)
		end := time.Since(start)
		mw.Metrix.Time.WithLabelValues(strconv.Itoa(mw.Response.Status), r.URL.Path).Observe(end.Seconds())

		mw.Metrix.Hits.WithLabelValues(strconv.Itoa(mw.Response.Status), r.URL.Path).Inc()
	})
}
