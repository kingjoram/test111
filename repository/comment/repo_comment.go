package comment

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"

	_ "github.com/jackc/pgx/stdlib"
)

type ICommentRepo interface {
	PingDb() error
	GetFilmRating(filmId uint64) (float64, error)
	GetFilmComments(filmId uint64, first uint64, last uint64) ([]CommentItem, error)
}

type RepoPostgre struct {
	DB *sql.DB
}

func GetCommentRepo(config configs.DbDsnCfg, lg *slog.Logger) *RepoPostgre {
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

func (repo *RepoPostgre) GetFilmRating(filmId uint64) (float64, error) {
	var rating float64
	err := repo.DB.QueryRow(
		"SELECT AVG(rating) FROM user_comment "+
			"WHERE id_film = $1", filmId).Scan(&rating)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return rating, nil
}

func (repo *RepoPostgre) GetFilmComments(filmId uint64, first uint64, last uint64) ([]CommentItem, error) {
	var comments []CommentItem

	rows, err := repo.DB.Query(
		"SELECT profile.login, rating, comment FROM users_comment "+
			"JOIN profile ON users_comment.id_user = profile.id "+
			"WHERE id_film = $1"+
			"OFFSET $2 LIMIT $3", filmId, first, last)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := CommentItem{}
		err := rows.Scan(&post.Username, &post.Rating, &post.Comment)
		if err != nil {
			return nil, err
		}
		comments = append(comments, post)
	}

	return comments, nil
}
