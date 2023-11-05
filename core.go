package main

import (
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/profile"
)

type Core struct {
	sessions map[string]Session
	mutex    sync.RWMutex
	lg       *slog.Logger
	Films    film.IFilmsRepo
	Users    profile.IUserRepo
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (core *Core) CreateSession(login string) (string, Session, error) {
	SID := RandStringRunes(32)

	session := Session{
		Login:     login,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	core.mutex.Lock()
	core.sessions[SID] = session
	core.mutex.Unlock()

	return SID, session, nil
}

func (core *Core) KillSession(sid string) error {
	core.mutex.Lock()
	delete(core.sessions, sid)
	core.mutex.Unlock()
	return nil
}

func (core *Core) FindActiveSession(sid string) (bool, error) {
	core.mutex.RLock()
	_, found := core.sessions[sid]
	core.mutex.RUnlock()
	return found, nil
}

func (core *Core) CreateUserAccount(request SignupRequest) {
	err := core.Users.CreateUser(request.Login, request.Password, request.Name, request.BirthDate, request.Email)
	if err != nil {
		core.lg.Error("create user error", "err", err.Error())
	}
}

func (core *Core) FindUserAccount(login string, password string) (*profile.UserItem, bool) {
	user, found, err := core.Users.GetUser(login, password)
	if err != nil {
		core.lg.Error("find user error", "err", err.Error())
	}
	return user, found
}

func (core *Core) FindUserByLogin(login string) bool {
	found, err := core.Users.FindUser(login)
	if err != nil {
		core.lg.Error("find user error", "err", err.Error())
	}

	return found
}

func RandStringRunes(seed int) string {
	symbols := make([]rune, seed)
	for i := range symbols {
		symbols[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(symbols)
}

func (core *Core) GetFilmsByGenre(genre string, start uint32, end uint32) []film.FilmItem {
	films, err := core.Films.GetFilmsByGenre(genre, start, end)
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
	}

	return films
}

func (core *Core) GetFilms(start uint32, end uint32) []film.FilmItem {
	films, err := core.Films.GetFilms(start, end)
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
	}

	return films
}
