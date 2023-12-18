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
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/metrics"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/middleware"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type IApi interface {
	SendResponse(w http.ResponseWriter, response requests.Response)
	Signin(w http.ResponseWriter, r *http.Request)
	SigninResponse(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	LogoutSession(w http.ResponseWriter, r *http.Request)
	AuthAccept(w http.ResponseWriter, r *http.Request)
}

type API struct {
	core usecase.ICore
	lg   *slog.Logger
	mx   *http.ServeMux
	mw   *middleware.ResponseMiddleware
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
    resp := &requests.Response{
        Status: http.StatusOK,
        Body:   nil,
    }
    middleware := &middleware.ResponseMiddleware{
        Response: resp,
		Metrix: metrics.GetMetrics(),
    }
    api := &API{
        core: c,
        lg:   l.With("module", "api"),
        mx:   http.NewServeMux(),
        mw:   middleware,
    }

	api.mx.Handle("/metrics", promhttp.Handler())
	api.mx.Handle("/signin", api.mw.GetResponse(http.HandlerFunc(api.Signin), l))
	api.mx.Handle("/signup", api.mw.GetResponse(http.HandlerFunc(api.Signup), l))
	api.mx.Handle("/logout", api.mw.GetResponse(http.HandlerFunc(api.LogoutSession), l))
	api.mx.Handle("/authcheck", api.mw.GetResponse(http.HandlerFunc(api.AuthAccept), l))
	api.mx.Handle("/api/v1/csrf", api.mw.GetResponse(http.HandlerFunc(api.GetCsrfToken), l))
	api.mx.Handle("/api/v1/settings", api.mw.GetResponse(http.HandlerFunc(api.Profile), l))

	return api
}

func (a *API) LogoutSession(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		a.mw.Response.Status = http.StatusUnauthorized
		return
	}

	found, _ := a.core.FindActiveSession(r.Context(), session.Value)
	if !found {
		a.mw.Response.Status = http.StatusUnauthorized
		return
	} else {
		err := a.core.KillSession(r.Context(), session.Value)
		if err != nil {
			a.lg.Error("failed to kill session", "err", err.Error())
		}
		session.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, session)
	}
}

func (a *API) AuthAccept(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	var authorized bool

	session, err := r.Cookie("session_id")
	if err == nil && session != nil {
		authorized, _ = a.core.FindActiveSession(r.Context(), session.Value)
	}

	if !authorized {
		a.mw.Response.Status = http.StatusUnauthorized
		return
	}
	login, err := a.core.GetUserName(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("auth accept error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	role, err := a.core.GetUserRole(login)
	if err != nil {
		a.lg.Error("auth accept error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	authCheckResponse := requests.AuthCheckResponse{Login: login, Role: role}
	a.mw.Response.Body = authCheckResponse
}

func (a *API) Signin(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		a.mw.Response.Status = http.StatusMethodNotAllowed
		return
	}

	csrfToken := r.Header.Get("x-csrf-token")

	_, err := a.core.CheckCsrfToken(r.Context(), csrfToken)
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		a.mw.Response.Status = http.StatusPreconditionFailed
		return
	}

	var request requests.SigninRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	user, found, err := a.core.FindUserAccount(request.Login, request.Password)
	if err != nil {
		a.lg.Error("Signin error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}
	if !found {
		a.mw.Response.Status = http.StatusUnauthorized
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
}

func (a *API) Signup(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method != http.MethodPost {
		a.mw.Response.Status = http.StatusMethodNotAllowed
		return
	}

	csrfToken := r.Header.Get("x-csrf-token")

	_, err := a.core.CheckCsrfToken(r.Context(), csrfToken)
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		a.mw.Response.Status = http.StatusPreconditionFailed
		return
	}

	var request requests.SignupRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.lg.Error("Signup error", "err", err.Error())
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		a.lg.Error("Signup error", "err", err.Error())
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	found, err := a.core.FindUserByLogin(request.Login)
	if err != nil {
		a.lg.Error("Signup error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	if found {
		a.mw.Response.Status = http.StatusConflict
		return
	}

	err = a.core.CreateUserAccount(request.Login, request.Password, request.Name, request.BirthDate, request.Email)
	if err == usecase.InvalideEmail {
		a.lg.Error("create user error", "err", err.Error())
		a.mw.Response.Status = http.StatusBadRequest
	}
	if err != nil {
		a.lg.Error("failed to create user account", "err", err.Error())
		a.mw.Response.Status = http.StatusBadRequest
	}
}

func (a *API) GetCsrfToken(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}

	csrfToken := r.Header.Get("x-csrf-token")

	found, err := a.core.CheckCsrfToken(r.Context(), csrfToken)
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}
	if csrfToken != "" && found {
		w.Header().Set("X-CSRF-Token", csrfToken)
		return
	}

	token, err := a.core.CreateCsrfToken(r.Context())
	if err != nil {
		w.Header().Set("X-CSRF-Token", "null")
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	w.Header().Set("X-CSRF-Token", token)
}

func (a *API) Profile(w http.ResponseWriter, r *http.Request) {
	a.mw.Response = &requests.Response{Status: http.StatusOK, Body: nil}
	if r.Method == http.MethodGet {
		session, err := r.Cookie("session_id")
		if err == http.ErrNoCookie {
			a.mw.Response.Status = http.StatusUnauthorized
			return
		}

		login, err := a.core.GetUserName(r.Context(), session.Value)
		if err != nil {
			a.lg.Error("Get Profile error", "err", err.Error())
		}

		profile, err := a.core.GetUserProfile(login)
		if err != nil {
			a.mw.Response.Status = http.StatusInternalServerError
			return
		}

		profileResponse := requests.ProfileResponse{
			Email:     profile.Email,
			Name:      profile.Name,
			Login:     profile.Login,
			Photo:     profile.Photo,
			BirthDate: profile.Birthdate,
		}

		a.mw.Response.Body = profileResponse
		return
	}

	if r.Method != http.MethodPost {
		a.mw.Response.Status = http.StatusUnauthorized
		return
	}
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		a.mw.Response.Status = http.StatusUnauthorized
		return
	}

	prevLogin, err := a.core.GetUserName(r.Context(), session.Value)
	if err != nil {
		a.lg.Error("Get Profile error", "err", err.Error())
	}

	err1 := r.ParseMultipartForm(10 << 20)
	if err1 != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	email := r.FormValue("email")
	login := r.FormValue("login")
	birthDate := r.FormValue("birthday")
	password := r.FormValue("password")
	photo, handler, err := r.FormFile("photo")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		a.lg.Error("Post profile error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	isRepeatPassword, err := a.core.CheckPassword(login, password)

	if isRepeatPassword {
		a.mw.Response.Status = http.StatusConflict
		return
	}

	var filename string
	if handler == nil {
		filename = ""

		err = a.core.EditProfile(prevLogin, login, password, email, birthDate, filename)
		if err != nil {
			a.lg.Error("Post profile error", "err", err.Error())
			a.mw.Response.Status = http.StatusInternalServerError
			return
		}
		return
	}

	filename = "/avatars/" + handler.Filename

	if err != nil && handler != nil && photo != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		a.mw.Response.Status = http.StatusBadRequest
		return
	}

	filePhoto, err := os.OpenFile("/home/ubuntu/frontend-project"+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}
	defer filePhoto.Close()

	_, err = io.Copy(filePhoto, photo)
	if err != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}

	err = a.core.EditProfile(prevLogin, login, password, email, birthDate, filename)
	if err != nil {
		a.lg.Error("Post profile error", "err", err.Error())
		a.mw.Response.Status = http.StatusInternalServerError
		return
	}
}
