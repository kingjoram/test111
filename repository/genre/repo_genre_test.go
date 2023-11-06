package genre

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetFilmGenres(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Title"})

	expect := []GenreItem{
		{Id: 1, Title: "t1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Title)
	}

	mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

	repo := &RepoPostgre{
		DB: db,
	}

	genres, err := repo.GetFilmGenres(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(genres, expect) {
		t.Errorf("results not match, want %v, have %v", expect, genres)
		return
	}

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	genres, err = repo.GetFilmGenres(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if genres != nil {
		t.Errorf("get comments error, comments should be nil")
	}
}
