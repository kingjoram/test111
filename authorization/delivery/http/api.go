package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/usecase"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
)

type IApi interface {
	SendResponse(w http.ResponseWriter, response requests.Response)
	Signin(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	LogoutSession(w http.ResponseWriter, r *http.Request)
	AuthAccept(w http.ResponseWriter, r *http.Request)
}

type API struct {
	core usecase.ICore
	lg   *slog.Logger
	mx   *http.ServeMux
}

func (a *API) ListenAndServe() error {
	err := http.ListenAndServe(":8081", a.mx)
	if err != nil {
		a.lg.Error("ListenAndServe error", "err", err.Error())
		return fmt.Errorf("listen and serve error: %w", err)
	}

	return nil
}

func GetApi(c *usecase.Core, l *slog.Logger) *API {
	api := &API{
		core: c,
		lg:   l.With("module", "api"),
		mx:   &http.ServeMux{},
	}
	mx := http.NewServeMux()
	mx.HandleFunc("/signup", api.Signup)
	mx.HandleFunc("/signin", api.Signin)
	mx.HandleFunc("/logout", api.LogoutSession)
	mx.HandleFunc("/authcheck", api.AuthAccept)
	mx.HandleFunc("/api/v1/csrf", api.GetCsrfToken)
	mx.HandleFunc("/api/v1/settings", api.Profile)

	api.mx = mx

	return api
}

func (a *API) LogoutSession(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}

	found, _ := a.core.FindActiveSession(r.Context(), session.Value)
	if !found {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	} else {
		err := a.core.KillSession(r.Context(), session.Value)
		if err != nil {
			a.lg.Error("failed to kill session", "err", err.Error())
		}
		session.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, session)
	}

	requests.SendResponse(w, response, a.lg)
}

func (a *API) AuthAccept(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	var authorized bool

	session, err := r.Cookie("session_id")
	if err == nil && session != nil {
		authorized, _ = a.core.FindActiveSession(r.Context(), session.Value)
	}

	if !authorized {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}
	login, err := a.core.GetUserName(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("auth accept error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	role, err := a.core.GetUserRole(login)
	if err != nil {
		a.lg.Error("auth accept error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	authCheckResponse := requests.AuthCheckResponse{Login: login, Role: role}
	response.Body = authCheckResponse

	requests.SendResponse(w, response, a.lg)
}

func (a *API) Signin(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	csrfToken := r.Header.Get("x-csrf-token")

	_, err := a.core.CheckCsrfToken(r.Context(), csrfToken)
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		response.Status = http.StatusPreconditionFailed
		requests.SendResponse(w, response, a.lg)
		return
	}

	var request requests.SigninRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	user, found, err := a.core.FindUserAccount(request.Login, request.Password)
	if err != nil {
		a.lg.Error("Signin error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}
	if !found {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	} else {
		sid, session, _ := a.core.CreateSession(r.Context(), user.Login)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    sid,
			Path:     "/",
			Expires:  session.ExpiresAt,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}

	requests.SendResponse(w, response, a.lg)
}

func (a *API) Signup(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		response.Status = http.StatusMethodNotAllowed
		requests.SendResponse(w, response, a.lg)
		return
	}

	csrfToken := r.Header.Get("x-csrf-token")

	_, err := a.core.CheckCsrfToken(r.Context(), csrfToken)
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		response.Status = http.StatusPreconditionFailed
		requests.SendResponse(w, response, a.lg)
		return
	}

	var request requests.SignupRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.lg.Error("Signup error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.lg.Error("Signup error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	found, err := a.core.FindUserByLogin(request.Login)
	if err != nil {
		a.lg.Error("Signup error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}
	if found {
		response.Status = http.StatusConflict
		requests.SendResponse(w, response, a.lg)
		return
	}
	err = a.core.CreateUserAccount(request.Login, request.Password, request.Name, request.BirthDate, request.Email)
	if err == usecase.InvalideEmail {
		a.lg.Error("create user error", "err", err.Error())
		response.Status = http.StatusBadRequest
	}
	if err != nil {
		a.lg.Error("failed to create user account", "err", err.Error())
		response.Status = http.StatusBadRequest
	}

	requests.SendResponse(w, response, a.lg)
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

func (a *API) Profile(w http.ResponseWriter, r *http.Request) {
	response := requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method == http.MethodGet {
		session, err := r.Cookie("session_id")
		if err == http.ErrNoCookie {
			response.Status = http.StatusUnauthorized
			requests.SendResponse(w, response, a.lg)
			return
		}

		login, err := a.core.GetUserName(r.Context(), session.Value)
		if err != nil {
			a.lg.Error("Get Profile error", "err", err.Error())
		}

		profile, err := a.core.GetUserProfile(login)
		if err != nil {
			response.Status = http.StatusInternalServerError
			requests.SendResponse(w, response, a.lg)
			return
		}

		profileResponse := requests.ProfileResponse{
			Email:     profile.Email,
			Name:      profile.Name,
			Login:     profile.Login,
			Photo:     profile.Photo,
			BirthDate: profile.Birthdate,
		}

		response.Body = profileResponse
		requests.SendResponse(w, response, a.lg)
		return
	}
	if r.Method != http.MethodPost {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response.Status = http.StatusUnauthorized
		requests.SendResponse(w, response, a.lg)
		return
	}

	prevLogin, err := a.core.GetUserName(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("Get Profile error", "err", err.Error())
	}

	err1 := r.ParseMultipartForm(10 << 20)
	if err1 != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}
	email := r.FormValue("email")
	login := r.FormValue("login")
	birthDate := r.FormValue("birthday")
	password := r.FormValue("password")
	photo, handler, err := r.FormFile("photo")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		a.lg.Error("Post profile error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	isRepeatPassword, err := a.core.CheckPassword(login, password)

	if isRepeatPassword {
		response.Status = http.StatusConflict
		requests.SendResponse(w, response, a.lg)
		return
	}

	var filename string
	if handler == nil {
		filename = ""

		err = a.core.EditProfile(prevLogin, login, password, email, birthDate, filename)
		if err != nil {
			a.lg.Error("Post profile error", "err", err.Error())
			response.Status = http.StatusInternalServerError
			requests.SendResponse(w, response, a.lg)
			return
		}
		requests.SendResponse(w, response, a.lg)
		return
	}

	filename = "/avatars/" + handler.Filename

	if err != nil && handler != nil && photo != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		response.Status = http.StatusBadRequest
		requests.SendResponse(w, response, a.lg)
		return
	}

	filePhoto, err := os.OpenFile("/home/ubuntu/frontend-project"+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}
	defer filePhoto.Close()

	_, err = io.Copy(filePhoto, photo)
	if err != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	err = a.core.EditProfile(prevLogin, login, password, email, birthDate, filename)
	if err != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		response.Status = http.StatusInternalServerError
		requests.SendResponse(w, response, a.lg)
		return
	}

	requests.SendResponse(w, response, a.lg)
}
