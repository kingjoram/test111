package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/errors"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/comment"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/crew"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/csrf"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/genre"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/profession"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/profile"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/session"
)

type Core struct {
	sessions   session.SessionRepo
	csrfTokens csrf.CsrfRepo
	mutex      sync.RWMutex
	lg         *slog.Logger
	films      film.IFilmsRepo
	users      profile.IUserRepo
	genres     genre.IGenreRepo
	comments   comment.ICommentRepo
	crew       crew.ICrewRepo
	profession profession.IProfessionRepo
}

type Session struct {
	Login     string
	ExpiresAt time.Time
}

func GetCore(cfg_sql configs.DbDsnCfg, cfg_csrf configs.DbRedisCfg, cfg_sessions configs.DbRedisCfg, lg *slog.Logger) (*Core, error) {
	csrf, err := csrf.GetCsrfRepo(cfg_csrf, lg)

	if err != nil {
		lg.Error("Csrf repository is not responding")
		return nil, err
	}
	session, err := session.GetSessionRepo(cfg_sessions, lg)

	if err != nil {
		lg.Error("Session repository is not responding")
		return nil, err
	}

	films, err := film.GetFilmRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	users, err := profile.GetUserRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	genres, err := genre.GetGenreRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	comments, err := comment.GetCommentRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	crew, err := crew.GetCrewRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	professions, err := profession.GetProfessionRepo(cfg_sql, lg)
	if err != nil {
		lg.Error("cant create repo")
		return nil, err
	}
	core := Core{
		sessions:   *session,
		csrfTokens: *csrf,
		lg:         lg.With("module", "core"),
		films:      films,
		users:      users,
		genres:     genres,
		comments:   comments,
		crew:       crew,
		profession: professions,
	}
	return &core, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (core *Core) CheckCsrfToken(ctx context.Context, token string) (bool, error) {
	core.mutex.RLock()
	found, err := core.csrfTokens.CheckActiveCsrf(ctx, token, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, err
}

func (core *Core) CreateCsrfToken(ctx context.Context) (string, error) {
	sid := RandStringRunes(32)

	core.mutex.Lock()
	csrfAdded, err := core.csrfTokens.AddCsrf(
		ctx,
		csrf.Csrf{
			SID:       sid,
			ExpiresAt: time.Now().Add(3 * time.Hour),
		},
		core.lg,
	)
	core.mutex.Unlock()

	if !csrfAdded && err != nil {
		return "", err
	}

	if !csrfAdded {
		return "", nil
	}

	return sid, nil
}

func (core *Core) GetUserName(ctx context.Context, sid string) (string, error) {
	core.mutex.RLock()
	login, err := core.sessions.GetUserLogin(ctx, sid, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return "", err
	}

	return login, nil
}

func (core *Core) CreateSession(ctx context.Context, login string) (string, session.Session, error) {
	sid := RandStringRunes(32)

	newSession := session.Session{
		Login:     login,
		SID:       sid,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	core.mutex.Lock()
	sessionAdded, err := core.sessions.AddSession(ctx, newSession, core.lg)
	core.mutex.Unlock()

	if !sessionAdded && err != nil {
		return "", session.Session{}, err
	}

	if !sessionAdded {
		return "", session.Session{}, nil
	}

	return sid, newSession, nil
}

func (core *Core) FindActiveSession(ctx context.Context, sid string) (bool, error) {
	core.mutex.RLock()
	found, err := core.sessions.CheckActiveSession(ctx, sid, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, nil
}

func (core *Core) KillSession(ctx context.Context, sid string) error {
	core.mutex.Lock()
	_, err := core.sessions.DeleteSession(ctx, sid, core.lg)
	core.mutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (core *Core) CreateUserAccount(login string, password string, name string, birthDate string, email string) error {
	if matched, _ := regexp.MatchString(`@`, email); !matched {
		return errors.InvalideEmail
	}
	err := core.users.CreateUser(login, password, name, birthDate, email)
	if err != nil {
		core.lg.Error("create user error", "err", err.Error())
		return fmt.Errorf("CreateUserAccount err: %w", err)
	}

	return nil
}

func (core *Core) FindUserAccount(login string, password string) (*profile.UserItem, bool, error) {
	user, found, err := core.users.GetUser(login, password)
	if err != nil {
		core.lg.Error("find user error", "err", err.Error())
		return nil, false, fmt.Errorf("FindUserAccount err: %w", err)
	}
	return user, found, nil
}

func (core *Core) FindUserByLogin(login string) (bool, error) {
	found, err := core.users.FindUser(login)
	if err != nil {
		core.lg.Error("find user error", "err", err.Error())
		return false, fmt.Errorf("FindUserByLogin err: %w", err)
	}

	return found, nil
}

func RandStringRunes(seed int) string {
	symbols := make([]rune, seed)
	for i := range symbols {
		symbols[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(symbols)
}

func (core *Core) GetFilmsByGenre(genre uint64, start uint64, end uint64) ([]film.FilmItem, error) {
	films, err := core.films.GetFilmsByGenre(genre, start, end)
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
		return nil, fmt.Errorf("GetFilmsByGenre err: %w", err)
	}

	return films, nil
}

func (core *Core) GetFilms(start uint64, end uint64) ([]film.FilmItem, error) {
	films, err := core.films.GetFilms(start, end)
	if err != nil {
		core.lg.Error("failed to get films from db", "err", err.Error())
		return nil, fmt.Errorf("GetFilms err: %w", err)
	}

	return films, nil
}

func (core *Core) GetFilm(filmId uint64) (*film.FilmItem, error) {
	film, err := core.films.GetFilm(filmId)
	if err != nil {
		core.lg.Error("Get Film error", "err", err.Error())
		return nil, fmt.Errorf("GetFilm err: %w", err)
	}

	return film, nil
}

func (core *Core) GetFilmGenres(filmId uint64) ([]genre.GenreItem, error) {
	genres, err := core.genres.GetFilmGenres(filmId)
	if err != nil {
		core.lg.Error("Get Film Genres error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmGenres err: %w", err)
	}

	return genres, nil
}

func (core *Core) GetFilmRating(filmId uint64) (float64, uint64, error) {
	rating, number, err := core.comments.GetFilmRating(filmId)
	if err != nil {
		core.lg.Error("Get Film Rating error", "err", err.Error())
		return 0, 0, fmt.Errorf("GetFilmRating err: %w", err)
	}

	return rating, number, nil
}

func (core *Core) GetFilmDirectors(filmId uint64) ([]crew.CrewItem, error) {
	directors, err := core.crew.GetFilmDirectors(filmId)
	if err != nil {
		core.lg.Error("Get Film Directors error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmDirectors err: %w", err)
	}

	return directors, nil
}

func (core *Core) GetFilmScenarists(filmId uint64) ([]crew.CrewItem, error) {
	scenarists, err := core.crew.GetFilmScenarists(filmId)
	if err != nil {
		core.lg.Error("Get Film Scenarists error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmScenarists err: %w", err)
	}

	return scenarists, nil
}

func (core *Core) GetFilmCharacters(filmId uint64) ([]crew.Character, error) {
	characters, err := core.crew.GetFilmCharacters(filmId)
	if err != nil {
		core.lg.Error("Get Film Characters error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmCharacters err: %w", err)
	}

	return characters, nil
}

func (core *Core) GetFilmComments(filmId uint64, first uint64, limit uint64) ([]comment.CommentItem, error) {
	comments, err := core.comments.GetFilmComments(filmId, first, limit)
	if err != nil {
		core.lg.Error("Get Film Comments error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmComments err: %w", err)
	}

	return comments, nil
}

func (core *Core) GetActor(actorId uint64) (*crew.CrewItem, error) {
	actor, err := core.crew.GetActor(actorId)
	if err != nil {
		core.lg.Error("Get Actor error", "err", err.Error())
		return nil, fmt.Errorf("GetActor err: %w", err)
	}

	return actor, nil
}

func (core *Core) GetActorsCareer(actorId uint64) ([]profession.ProfessionItem, error) {
	career, err := core.profession.GetActorsProfessions(actorId)
	if err != nil {
		core.lg.Error("Get Actors Career error", "err", err.Error())
		return nil, fmt.Errorf("GetActorsCareer err: %w", err)
	}

	return career, nil
}

func (core *Core) AddComment(filmId uint64, userLogin string, rating uint16, text string) error {
	err := core.comments.AddComment(filmId, userLogin, rating, text)
	if err != nil {
		core.lg.Error("Add Comment error", "err", err.Error())
		return fmt.Errorf("GetActorsCareer err: %w", err)
	}

	return nil
}

func (core *Core) GetUserProfile(login string) (*profile.UserItem, error) {
	profile, err := core.users.GetUserProfile(login)
	if err != nil {
		core.lg.Error("GetUserProfile error", "err", err.Error())
		return nil, fmt.Errorf("GetUserProfile err: %w", err)
	}

	return profile, nil
}

func (core *Core) GetGenre(genreId uint64) (string, error) {
	genre, err := core.genres.GetGenreById(genreId)
	if err != nil {
		core.lg.Error("GetGenre error", "err", err.Error())
		return "", fmt.Errorf("GetGenre err: %w", err)
	}

	return genre, nil
}

func (core *Core) EditProfile(prevLogin string, login string, password string, email string, birthDate string, photo string) error {
	err := core.users.EditProfile(prevLogin, login, password, email, birthDate, photo)
	if err != nil {
		core.lg.Error("Edit profile error", "err", err.Error())
		return fmt.Errorf("Edit profile error: %w", err)
	}

	return nil
}

func (core *Core) FindUsersComment(login string, filmId uint64) (bool, error) {
	found, err := core.comments.FindUsersComment(login, filmId)
	if err != nil {
		core.lg.Error("find users comment error", "err", err.Error())
		return false, fmt.Errorf("find users comment error: %w", err)
	}

	return found, nil
}
