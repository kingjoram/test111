package crew

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetFilmDirectors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Name", "Photo"})

	expect := []CrewItem{
		{Id: 1, Name: "n1", Photo: "p1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Name, item.Photo)
	}

	mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

	repo := &RepoPostgre{
		DB: db,
	}

	directors, err := repo.GetFilmDirectors(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(directors, expect) {
		t.Errorf("results not match, want %v, have %v", expect, directors)
		return
	}

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	directors, err = repo.GetFilmDirectors(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if directors != nil {
		t.Errorf("get comments error, comments should be nil")
	}
}

func TestGetFilmScenarists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Name", "Photo"})

	expect := []CrewItem{
		{Id: 1, Name: "n1", Photo: "p1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Name, item.Photo)
	}

	mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

	repo := &RepoPostgre{
		DB: db,
	}

	scenarists, err := repo.GetFilmScenarists(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(scenarists, expect) {
		t.Errorf("results not match, want %v, have %v", expect, scenarists)
		return
	}

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	scenarists, err = repo.GetFilmScenarists(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if scenarists != nil {
		t.Errorf("get comments error, comments should be nil")
	}
}

func TestGetFilmCharacters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Name", "Photo", "CharacterName"})

	expect := []Character{
		{IdActor: 1, NameActor: "n1", ActorPhoto: "p1", NameCharacter: "chn1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.IdActor, item.NameActor, item.ActorPhoto, item.NameCharacter)
	}

	mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

	repo := &RepoPostgre{
		DB: db,
	}

	characters, err := repo.GetFilmCharacters(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(characters, expect) {
		t.Errorf("results not match, want %v, have %v", expect, characters)
		return
	}

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	characters, err = repo.GetFilmCharacters(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if characters != nil {
		t.Errorf("get comments error, comments should be nil")
	}
}
