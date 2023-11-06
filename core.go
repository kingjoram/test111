package main

import (
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/comment"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/crew"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/genre"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/profile"
)

type Core struct {
	sessions map[string]Session
	mutex    sync.RWMutex
	lg       *slog.Logger
	Films    film.IFilmsRepo
	Users    profile.IUserRepo
	Genres   genre.IGenreRepo
	Comments comment.ICommentRepo
	Crew     crew.ICrewRepo
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

func (core *Core) GetFilmsByGenre(genre string, start uint64, end uint64) []film.FilmItem {
	films, err := core.Films.GetFilmsByGenre(genre, start, end)
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
	}

	return films
}

func (core *Core) GetFilms(start uint64, end uint64) []film.FilmItem {
	films, err := core.Films.GetFilms(start, end)
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
	}

	return films
}

func (core *Core) PingRepos(timer uint32) {
	for {
		err := core.Users.PingDb()
		if err != nil {
			core.lg.Error("Ping User repo error", "err", err.Error())
			return
		}
		err = core.Films.PingDb()
		if err != nil {
			core.lg.Error("Ping Film repo error", "err", err.Error())
			return
		}
		err = core.Genres.PingDb()
		if err != nil {
			core.lg.Error("Ping Genre repo error", "err", err.Error())
			return
		}
		err = core.Comments.PingDb()
		if err != nil {
			core.lg.Error("Ping Comment repo error", "err", err.Error())
			return
		}
		err = core.Comments.PingDb()
		if err != nil {
			core.lg.Error("Ping Crew repo error", "err", err.Error())
			return
		}

		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func (core *Core) GetFilm(filmId uint64) *film.FilmItem {
	film, err := core.Films.GetFilm(filmId)
	if err != nil {
		core.lg.Error("Get Film error", "err", err.Error())
	}

	return film
}

func (core *Core) GetFilmGenres(filmId uint64) []genre.GenreItem {
	genres, err := core.Genres.GetFilmGenres(filmId)
	if err != nil {
		core.lg.Error("Get Film Genres error", "err", err.Error())
	}

	return genres
}

func (core *Core) GetFilmRating(filmId uint64) float64 {
	rating, err := core.Comments.GetFilmRating(filmId)
	if err != nil {
		core.lg.Error("Get Film Rating error", "err", err.Error())
	}

	return rating
}

func (core *Core) GetFilmDirectors(filmId uint64) []crew.CrewItem {
	directors, err := core.Crew.GetFilmDirectors(filmId)
	if err != nil {
		core.lg.Error("Get Film Directors error", "err", err.Error())
	}

	return directors
}

func (core *Core) GetFilmScenarists(filmId uint64) []crew.CrewItem {
	scenarists, err := core.Crew.GetFilmScenarists(filmId)
	if err != nil {
		core.lg.Error("Get Film Scenarists error", "err", err.Error())
	}

	return scenarists
}

func (core *Core) GetFilmCharacters(filmId uint64) []crew.Character {
	characters, err := core.Crew.GetFilmCharacters(filmId)
	if err != nil {
		core.lg.Error("Get Film Characters error", "err", err.Error())
	}

	return characters
}

func (core *Core) GetFilmComments(filmId uint64, first uint64, last uint64) []comment.CommentItem {
	comments, err := core.Comments.GetFilmComments(filmId, first, last)
	if err != nil {
		core.lg.Error("Get Film Comments error", "err", err.Error())
	}

	return comments
}
