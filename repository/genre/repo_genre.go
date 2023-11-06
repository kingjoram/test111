package genre

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"

	_ "github.com/jackc/pgx/stdlib"
)

type IGenreRepo interface {
	PingDb() error
	GetFilmGenres(filmId uint64) ([]GenreItem, error)
}

type RepoPostgre struct {
	DB *sql.DB
}

func GetGenreRepo(config configs.DbDsnCfg, lg *slog.Logger) *RepoPostgre {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		config.User, config.DbName, config.Password, config.Host, config.Port, config.Sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		lg.Error("sql open error", "err", err.Error())
		return nil
	}
	err = db.Ping()
	if err != nil {
		lg.Error("sql ping error", "err", err.Error())
		return nil
	}
	db.SetMaxOpenConns(config.MaxOpenConns)

	postgreDb := RepoPostgre{DB: db}
	return &postgreDb
}

func (repo *RepoPostgre) PingDb() error {
	err := repo.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (repo *RepoPostgre) GetFilmGenres(filmId uint64) ([]GenreItem, error) {
	var genres []GenreItem

	rows, err := repo.DB.Query(
		"SELECT genre.id, genre.title FROM genre "+
			"JOIN films_genre ON genre.id = films_genre.id_genre "+
			"WHERE films_genre.id_film = $1", filmId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := GenreItem{}
		err := rows.Scan(&post.Id, &post.Title)
		if err != nil {
			return nil, err
		}
		genres = append(genres, post)
	}

	return genres, nil
}
