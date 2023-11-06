package profile

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
)

type IUserRepo interface {
	GetUser(login string, password string) (*UserItem, bool, error)
	FindUser(login string) (bool, error)
	CreateUser(login string, password string, name string, birthDate string, email string) error
	PingDb() error
}

type RepoPostgre struct {
	DB *sql.DB
}

func GetUserRepo(config configs.DbDsnCfg, lg *slog.Logger) *RepoPostgre {
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

func (repo *RepoPostgre) GetUser(login string, password string) (*UserItem, bool, error) {
	post := &UserItem{}

	err := repo.DB.QueryRow(
		"SELECT login, photo FROM profiles "+
			"WHERE login = $1 AND password = $2", login, password).Scan(&post.Login, &post.Photo)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return post, true, nil
}

func (repo *RepoPostgre) FindUser(login string) (bool, error) {
	post := &UserItem{}

	err := repo.DB.QueryRow(
		"SELECT login FROM profiles "+
			"WHERE login = $1", login).Scan(&post.Login)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (repo *RepoPostgre) CreateUser(login string, password string, name string, birthDate string, email string) error {
	_, err := repo.DB.Exec(
		"INSERT INTO profiles(name, birth_date, photo, login, password, email, registration_date) "+
			"VALUES($1, $2, '../../user_avatars/default.jpg', $3, $4, $5, CURRENT_TIMESTAMP)",
		name, birthDate, login, password, email)
	if err != nil {
		return err
	}

	return nil
}

func (repo *RepoPostgre) PingDb() error {
	err := repo.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}
