package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/crew"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/genre"
)

type API struct {
	core *Core
	lg   *slog.Logger
}

type Session struct {
	Login     string
	ExpiresAt time.Time
}

type FilmsResponse struct {
	Page           uint64          `json:"current_page"`
	PageSize       uint64          `json:"page_size"`
	CollectionName string          `json:"collection_name"`
	Total          uint64          `json:"total"`
	Films          []film.FilmItem `json:"films"`
}

type FilmResponse struct {
	Film       film.FilmItem     `json:"film"`
	Genres     []genre.GenreItem `json:"genres"`
	Rating     float64           `json:"rating"`
	Directors  []crew.CrewItem   `json:"directors"`
	Scenarists []crew.CrewItem   `json:"scenarists"`
	Characters []crew.Character  `json:"characters"`
}

func (a *API) SendResponse(w http.ResponseWriter, response Response) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.lg.Error("failed to pack json", "err", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResponse)
	if err != nil {
		a.lg.Error("failed to send response", "err", err.Error())
	}
}

func (a *API) Films(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}

	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		a.SendResponse(w, response)
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

	var films []film.FilmItem
	collectionId := r.URL.Query().Get("collection_id")
	if collectionId == "" {
		films = a.core.GetFilms(uint64((page-1)*pageSize), pageSize)
	} else {
		films = a.core.GetFilmsByGenre(collectionId, uint64((page-1)*pageSize), pageSize)
	}
	filmsResponse := FilmsResponse{
		Page:           page,
		PageSize:       pageSize,
		Total:          uint64(len(films)),
		CollectionName: collectionId,
		Films:          films,
	}
	response.Body = filmsResponse

	a.SendResponse(w, response)
}

func (a *API) LogoutSession(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
		a.SendResponse(w, response)
		return
	}

	found, _ := a.core.FindActiveSession(session.Value)
	if !found {
		response.Status = http.StatusUnauthorized
		a.SendResponse(w, response)
		return
	} else {
		err := a.core.KillSession(session.Value)
		if err != nil {
			a.core.lg.Error("failed to kill session", "err", err.Error())
		}
		session.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, session)
	}

	a.SendResponse(w, response)
}

func (a *API) AuthAccept(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	var authorized bool

	session, err := r.Cookie("session_id")
	if err == nil && session != nil {
		authorized, _ = a.core.FindActiveSession(session.Value)
	}

	if !authorized {
		response.Status = http.StatusUnauthorized
		a.SendResponse(w, response)
		return
	}

	a.SendResponse(w, response)
}

func (a *API) Signin(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		a.SendResponse(w, response)
		return
	}
	var request SigninRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Status = http.StatusBadRequest
		a.SendResponse(w, response)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		response.Status = http.StatusBadRequest
		a.SendResponse(w, response)
		return
	}

	user, found := a.core.FindUserAccount(request.Login, request.Password)
	if !found {
		response.Status = http.StatusUnauthorized
		a.SendResponse(w, response)
		return
	} else {
		sid, session, _ := a.core.CreateSession(user.Login)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    sid,
			Path:     "/",
			Expires:  session.ExpiresAt,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}

	a.SendResponse(w, response)
}

func (a *API) Signup(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		a.SendResponse(w, response)
		return
	}
	var request SignupRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Status = http.StatusBadRequest
		a.SendResponse(w, response)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		response.Status = http.StatusBadRequest
		a.SendResponse(w, response)
		return
	}

	found := a.core.FindUserByLogin(request.Login)
	if found {
		response.Status = http.StatusConflict
		a.SendResponse(w, response)
		return
	} else {
		a.core.CreateUserAccount(request)
		if err != nil {
			a.core.lg.Error("failed to create user account", "err", err.Error())
		}
	}

	a.SendResponse(w, response)
}

func (a *API) Film(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
		a.SendResponse(w, response)
		return
	}

	filmId, err := strconv.ParseUint(r.URL.Query().Get("film_id"), 10, 64)
	if err != nil {
		response.Status = http.StatusBadRequest
		a.SendResponse(w, response)
		return
	}

	film := a.core.GetFilm(filmId)
	if film == nil {
		response.Status = http.StatusNotFound
		a.SendResponse(w, response)
		return
	}
	genres := a.core.GetFilmGenres(filmId)
	rating := a.core.GetFilmRating(filmId)
	directors := a.core.GetFilmDirectors(filmId)
	scenarists := a.core.GetFilmScenarists(filmId)
	characters := a.core.GetFilmCharacters(filmId)

	filmResponse := FilmResponse{
		Film:       *film,
		Genres:     genres,
		Rating:     rating,
		Directors:  directors,
		Scenarists: scenarists,
		Characters: characters,
	}
	response.Body = filmResponse

	a.SendResponse(w, response)
}
