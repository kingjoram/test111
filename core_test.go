package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/profile"
)

func TestGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"photo", "login"})

	testUser := profile.UserItem{
		Photo: "url1",
		Login: "l1",
	}
	expect := []*profile.UserItem{&testUser}

	for _, item := range expect {
		rows = rows.AddRow(item.Login, item.Photo)
	}

	mock.ExpectQuery("SELECT login, photo FROM profiles WHERE").WithArgs(expect[0].Login, expect[0].Password).WillReturnRows(rows)

	repo := &profile.RepoPostgre{
		DB: db,
	}

	user, foundAccount, err := repo.GetUser(expect[0].Login, expect[0].Password)
	if err != nil {
		t.Errorf("GetUser error: %s", err)
	}
	if !foundAccount {
		t.Errorf("user not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(user, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], user)
		return
	}

	mock.
		ExpectQuery("SELECT login, photo FROM profiles WHERE").
		WithArgs(expect[0].Login, expect[0].Password).
		WillReturnError(fmt.Errorf("db_error"))

	_, found, err := repo.GetUser(expect[0].Login, expect[0].Password)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if found {
		t.Errorf("expected not found")
	}
}

func TestFindUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"login"})

	testUser := profile.UserItem{
		Login: "l1",
	}
	expect := []*profile.UserItem{&testUser}

	for _, item := range expect {
		rows = rows.AddRow(item.Login)
	}

	mock.ExpectQuery("SELECT login FROM profiles WHERE").WithArgs(expect[0].Login).WillReturnRows(rows)

	repo := &profile.RepoPostgre{
		DB: db,
	}

	foundAccount, err := repo.FindUser(expect[0].Login)
	if err != nil {
		t.Errorf("GetUser error: %s", err)
	}
	if !foundAccount {
		t.Errorf("user not found")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	mock.
		ExpectQuery("SELECT login FROM profiles WHERE").
		WithArgs(expect[0].Login).
		WillReturnError(fmt.Errorf("db_error"))

	found, err := repo.FindUser(expect[0].Login)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if found {
		t.Errorf("expected not found")
	}
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"login"})

	testUser := profile.UserItem{
		Login:     "l1",
		Password:  "p1",
		Birthdate: "2003-10-08",
		Name:      "n1",
		Email:     "e1",
	}
	expect := []*profile.UserItem{&testUser}

	for _, item := range expect {
		rows = rows.AddRow(item.Login)
	}

	mock.
		ExpectExec("INSERT INTO profiles").
		WithArgs(testUser.Name, testUser.Birthdate, testUser.Login, testUser.Password, testUser.Email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := &profile.RepoPostgre{
		DB: db,
	}

	err = repo.CreateUser(testUser.Login, testUser.Password, testUser.Name, testUser.Birthdate, testUser.Email)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	mock.
		ExpectExec("INSERT INTO profiles").
		WithArgs(testUser.Name, testUser.Birthdate, testUser.Login, testUser.Password, testUser.Email).
		WillReturnError(fmt.Errorf("db_error"))

	err = repo.CreateUser(testUser.Login, testUser.Password, testUser.Name, testUser.Birthdate, testUser.Email)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

}

func TestGetFilmsByGenre(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Title", "Poster"})

	expect := []film.FilmItem{
		{Id: 1, Title: "t1", Poster: "url1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Title, item.Poster)
	}

	mock.ExpectQuery("SELECT film.id, film.title, poster FROM film JOIN").WithArgs("g1", 1, 2).WillReturnRows(rows)

	repo := &film.RepoPostgre{
		DB: db,
	}

	films, err := repo.GetFilmsByGenre("g1", 1, 2)
	if err != nil {
		t.Errorf("GetFilmsByGenre error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(films, expect) {
		t.Errorf("results not match, want %v, have %v", expect, films)
		return
	}

	mock.
		ExpectQuery("SELECT film.id, film.title, poster FROM film JOIN").
		WithArgs("g3", 1, 2).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetFilmsByGenre("g3", 1, 2)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestGetFilms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Id", "Title", "Poster"})

	expect := []film.FilmItem{
		{Id: 1, Title: "t1", Poster: "url1"},
	}

	for _, item := range expect {
		rows = rows.AddRow(item.Id, item.Title, item.Poster)
	}

	mock.ExpectQuery("SELECT film.id, film.title, poster FROM film").WithArgs(1, 2).WillReturnRows(rows)

	repo := &film.RepoPostgre{
		DB: db,
	}

	films, err := repo.GetFilms(1, 2)
	if err != nil {
		t.Errorf("GetFilmsByGenre error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if !reflect.DeepEqual(films, expect) {
		t.Errorf("results not match, want %v, have %v", expect, films)
		return
	}

	mock.
		ExpectQuery("SELECT film.id, film.title, poster FROM film").
		WithArgs(1, 2).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetFilms(1, 2)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestCreateAndKillSession(t *testing.T) {
	login := "testLogin"
	testCore := Core{sessions: make(map[string]Session)}

	sid, _, _ := testCore.CreateSession(login)
	isFound, _ := testCore.FindActiveSession(sid)
	if !isFound {
		t.Errorf("session not found")
	}

	err := testCore.KillSession(sid)
	if err != nil {
		t.Errorf("failed to kill session")
	}

	isFound, _ = testCore.FindActiveSession(sid)
	if isFound {
		t.Errorf("found killed session")
	}
}

func TestFilmsPost(t *testing.T) {
	h := httptest.NewRequest(http.MethodPost, "/api/v1/films", nil)
	w := httptest.NewRecorder()

	api := API{}
	api.Films(w, h)
	var response Response

	body, _ := io.ReadAll(w.Body)
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Error("cant unmarshal jsone")
	}

	if response.Status != http.StatusMethodNotAllowed {
		t.Errorf("got incorrect status")
	}
}

func TestSignupGet(t *testing.T) {
	h := httptest.NewRequest(http.MethodGet, "/signup", nil)
	w := httptest.NewRecorder()

	api := API{}
	api.Signup(w, h)
	var response Response

	body, _ := io.ReadAll(w.Body)
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Error("cant unmarshal jsone")
	}

	if response.Status != http.StatusMethodNotAllowed {
		t.Errorf("got incorrect status")
	}
}

func TestSigninGet(t *testing.T) {
	h := httptest.NewRequest(http.MethodGet, "/signin", nil)
	w := httptest.NewRecorder()

	api := API{}
	api.Signin(w, h)
	var response Response

	body, _ := io.ReadAll(w.Body)
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Error("cant unmarshal jsone")
	}

	if response.Status != http.StatusMethodNotAllowed {
		t.Errorf("got incorrect status")
	}
}
