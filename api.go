package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
)

type API struct {
	core *Core
	lg   *slog.Logger
}

type Session struct {
	Login     string
	ExpiresAt time.Time
}

type Film struct {
	Title    string   `json:"title"`
	ImageURL string   `json:"imagine_url"`
	Rating   float64  `json:"rating"`
	Genres   []string `json:"genres"`
}

type FilmsResponse struct {
	Page           uint64 `json:"current_page"`
	CollectionName string `json:"collection_name"`
	Total          uint64 `json:"total"`
	Films          []Film `json:"films"`
}

func (a *API) Films(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodGet {
		response.Status = http.StatusMethodNotAllowed
	} else {
		collectionId := r.URL.Query().Get("collection_id")
		if collectionId == "" {
			collectionId = "new"
		}

		page, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
		if err != nil {
			page = 1
		}
		pageSize := uint64(8)

		collectionName, found := a.core.GetCollection(collectionId)
		if !found {
			collectionName = "Новинки"
		}

		films := GetFilms()
		if collectionName != "Новинки" {
			films = SortFilms(collectionName, films)
		}

		if uint64(cap(films)) < page*pageSize {
			page = uint64(math.Ceil(float64(uint64(cap(films)) / pageSize)))
		}
		filmsResponse := FilmsResponse{
			Page:           page,
			Total:          uint64(len(films)),
			CollectionName: collectionName,
			Films:          films[pageSize*(page-1) : pageSize*page],
		}
		response.Body = filmsResponse
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.lg.Error("failed to pack json", "err", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (a *API) LogoutSession(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
	}

	found := a.core.FindActiveSession(session.Value)
	if !found {
		response.Status = http.StatusUnauthorized
	} else {
		a.core.KillSession(session.Value)
		session.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, session)
	}

	answer, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}

func (a *API) Signin(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
	} else {
		var request SigninRequest

		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.Status = http.StatusBadRequest
		}

		if err = json.Unmarshal(body, &request); err != nil {
			response.Status = http.StatusBadRequest
		}

		user, found := a.core.FindUserAccount(request.Login)
		if !found || user.Password != request.Password {
			response.Status = http.StatusUnauthorized
		} else {
			sid, session := a.core.CreateSession(user.Login)
			cookie := &http.Cookie{
				Name:     "session_id",
				Value:    sid,
				Path:     "/",
				Expires:  session.ExpiresAt,
				HttpOnly: false,
			}
			http.SetCookie(w, cookie)
		}
	}

	answer, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}

func (a *API) Signup(w http.ResponseWriter, r *http.Request) {
	response := Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
	} else {
		var request SignupRequest

		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.Status = http.StatusBadRequest
		}

		err = json.Unmarshal(body, &request)
		if err != nil {
			response.Status = http.StatusBadRequest
		}

		_, found := a.core.FindUserAccount(request.Login)
		if found {
			response.Status = http.StatusConflict
		} else {
			a.core.CreateUserAccount(request)
		}
	}

	answer, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}
