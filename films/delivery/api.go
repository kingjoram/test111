package delivery

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/usecase"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
)

type API struct {
	core   usecase.ICore
	lg     *slog.Logger
	mx     *http.ServeMux
	adress string
}

func GetApi(c *usecase.Core, l *slog.Logger, cfg *configs.DbDsnCfg) *API {
	api := &API{
		core:   c,
		lg:     l.With("module", "api"),
		adress: cfg.ServerAdress,
	}
	mx := http.NewServeMux()
	mx.HandleFunc("/api/v1/films", api.Films)
	mx.HandleFunc("/api/v1/film", api.Film)
	mx.HandleFunc("/api/v1/actor", api.Actor)
	mx.HandleFunc("/api/v1/favorite/films", api.FavoriteFilms)
	mx.HandleFunc("/api/v1/favorite/film/add", api.FavoriteFilmsAdd)
	mx.HandleFunc("/api/v1/favorite/film/remove", api.FavoriteFilmsRemove)
	mx.HandleFunc("/api/v1/favorite/actors", api.FavoriteActors)
	mx.HandleFunc("/api/v1/favorite/actor/add", api.FavoriteActorsAdd)
	mx.HandleFunc("/api/v1/favorite/actor/remove", api.FavoriteActorsRemove)
	mx.HandleFunc("/api/v1/find", api.FindFilm)
	mx.HandleFunc("/api/v1/search/actor", api.FindActor)
	mx.HandleFunc("/api/v1/calendar", api.Calendar)

	api.mx = mx

	return api
}

func (a *API) ListenAndServe() {
	err := http.ListenAndServe(a.adress, a.mx)
	if err != nil {
		a.lg.Error("listen and serve error", "err", err.Error())
	}
}

func (a *API) Films(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}

	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	page, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.ParseUint(r.URL.Query().Get("page_size"), 10, 64)
	if err != nil {
		pageSize = 8
	}

	genreId, err := strconv.ParseUint(r.URL.Query().Get("collection_id"), 10, 64)
	if err != nil {
		genreId = 0
	}

	var films []models.FilmItem

	films, genre, err := a.core.GetFilmsAndGenreTitle(genreId, uint64((page-1)*pageSize), pageSize)
	if err != nil {
		a.lg.Error("get films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	filmsResponse := requests.FilmsResponse{
		Page:           page,
		PageSize:       pageSize,
		Total:          uint64(len(films)),
		CollectionName: genre,
		Films:          films,
	}
	response.Body = filmsResponse

	requests.SendResponse(w, response, a.lg)
}

func (a *API) Film(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	filmId, err := strconv.ParseUint(r.URL.Query().Get("film_id"), 10, 64)
	if err != nil {
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	film, err := a.core.GetFilmInfo(filmId)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("film error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = film

	requests.SendResponse(w, response, a.lg)
}

func (a *API) Actor(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	actorId, err := strconv.ParseUint(r.URL.Query().Get("actor_id"), 10, 64)
	if err != nil {
		a.lg.Error("actor error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	actor, err := a.core.GetActorInfo(actorId)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("actor error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = actor
	requests.SendResponse(w, response, a.lg)
}

func (a *API) FindFilm(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
	var request requests.FindFilmRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	films, err := a.core.FindFilm(request.Title, request.DateFrom, request.DateTo, request.RatingFrom, request.RatingTo, request.Mpaa, request.Genres, request.Actors)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("find film error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	filmsResponse := requests.FilmsResponse{
		Films: films,
		Total: uint64(len((films))),
	}
	response.Body = filmsResponse
	requests.SendResponse(w, response, a.lg)
}

func (a *API) FavoriteFilmsAdd(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	userId, err := a.core.GetUserId(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	filmId, err := strconv.ParseUint(r.URL.Query().Get("film_id"), 10, 64)
	if err != nil {
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	err = a.core.FavoriteFilmsAdd(userId, filmId)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusBadRequest
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	requests.SendResponse(w, response, a.lg)
}

func (a *API) FavoriteFilmsRemove(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	userId, err := a.core.GetUserId(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	filmId, err := strconv.ParseUint(r.URL.Query().Get("film_id"), 10, 64)
	if err != nil {
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	err = a.core.FavoriteFilmsRemove(userId, filmId)
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	requests.SendResponse(w, response, a.lg)
}

func (a *API) FavoriteFilms(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	userId, err := a.core.GetUserId(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	page, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.ParseUint(r.URL.Query().Get("per_page"), 10, 64)
	if err != nil {
		pageSize = 8
	}

	films, err := a.core.FavoriteFilms(userId, uint64((page-1)*pageSize), pageSize)
	if err != nil {
		a.lg.Error("favorite films error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = films
	requests.SendResponse(w, response, a.lg)
}

func (a *API) FavoriteActorsAdd(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteActorsRemove(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) FavoriteActors(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
}

func (a *API) Calendar(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	calendar, err := a.core.GetCalendar()
	if err != nil {
		a.lg.Error("calendar error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = calendar
	requests.SendResponse(w, response, a.lg)
}

func (a *API) FindActor(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}
	var request requests.FindActorRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.lg.Error("find actor error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		a.lg.Error("find actor error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	actors, err := a.core.FindActor(request.Name, request.BirthDate, request.Films, request.Career, request.Country)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			response.Status = http.StatusNotFound
			requests.SendResponse(w, response, a.lg)
			return
		}
		a.lg.Error("find actor error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	response.Body = actors
	requests.SendResponse(w, response, a.lg)
}
