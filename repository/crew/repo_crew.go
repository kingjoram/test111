package crew

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"

	_ "github.com/jackc/pgx/stdlib"
)

type ICrewRepo interface {
	PingDb() error
	GetFilmDirectors(filmId uint64) ([]CrewItem, error)
	GetFilmScenarists(filmId uint64) ([]CrewItem, error)
	GetFilmCharacters(filmId uint64) ([]Character, error)
}

type RepoPostgre struct {
	DB *sql.DB
}

func GetCrewRepo(config configs.DbDsnCfg, lg *slog.Logger) *RepoPostgre {
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

func (repo *RepoPostgre) GetFilmDirectors(filmId uint64) ([]CrewItem, error) {
	var directors []CrewItem

	rows, err := repo.DB.Query(
		"SELECT crew.id, name, photo  FROM crew "+
			"JOIN person_in_film ON crew.id = person_in_film.id_person "+
			"WHERE id_film = $1 AND id_profession = "+
			"(SELECT id FROM profession WHERE title = 'режиссёр')", filmId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := CrewItem{}
		err := rows.Scan(&post.Id, &post.Name, &post.Photo)
		if err != nil {
			return nil, err
		}
		directors = append(directors, post)
	}

	return directors, nil
}

func (repo *RepoPostgre) GetFilmScenarists(filmId uint64) ([]CrewItem, error) {
	var scenarists []CrewItem

	rows, err := repo.DB.Query(
		"SELECT crew.id, name, photo  FROM crew "+
			"JOIN person_in_film ON crew.id = person_in_film.id_person "+
			"WHERE id_film = $1 AND id_profession = "+
			"(SELECT id FROM profession WHERE title = 'сценарист')", filmId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := CrewItem{}
		err := rows.Scan(&post.Id, &post.Name, &post.Photo)
		if err != nil {
			return nil, err
		}
		scenarists = append(scenarists, post)
	}

	return scenarists, nil
}

func (repo *RepoPostgre) GetFilmCharacters(filmId uint64) ([]Character, error) {
	characters := []Character{}

	rows, err := repo.DB.Query(
		"SELECT crew.id, name, photo, person_in_film.character_name FROM crew "+
			"JOIN person_in_film ON crew.id = person_in_film.id_person "+
			"WHERE id_film = $1 AND id_profession = "+
			"(SELECT id FROM profession WHERE title = 'актёр')", filmId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := Character{}
		err := rows.Scan(&post.IdActor, &post.NameActor, &post.ActorPhoto, &post.NameCharacter)
		if err != nil {
			return nil, err
		}
		characters = append(characters, post)
	}

	return characters, nil
}
