package comment

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetFilmRating(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Average"})

	expect := 4.2

	rows = rows.AddRow(expect)

	mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

	repo := &RepoPostgre{
		DB: db,
	}

	rating, err := repo.GetFilmRating(1)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if rating != expect {
		t.Errorf("results not match, want %v, have %v", expect, rating)
		return
	}

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetFilmRating(1)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestGetFilmComments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Login", "Rating", "Comment"})

	expect := []CommentItem{
		{Username: "l1", Rating: 4, Comment: "c1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Username, item.Rating, item.Comment)
	}

	mock.ExpectQuery("SELECT").WithArgs(1, 0, 5).WillReturnRows(rows)

	repo := &RepoPostgre{
		DB: db,
	}

	comments, err := repo.GetFilmComments(1, 0, 5)
	if err != nil {
		t.Errorf("GetFilm error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(comments, expect) {
		t.Errorf("results not match, want %v, have %v", expect, comments)
		return
	}

	mock.
		ExpectQuery("SELECT").
		WithArgs(1, 0, 5).
		WillReturnError(fmt.Errorf("db_error"))

	comments, err = repo.GetFilmComments(1, 0, 5)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if comments != nil {
		t.Errorf("get comments error, comments should be nil")
	}
}
