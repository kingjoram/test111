package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCreateUserAccount(t *testing.T) {
	login := "testLogin"
	email := "test@mail.ru"
	password := "testPassword"
	testCore := Core{users: make(map[string]User)}
	testRequest := SignupRequest{
		Login:    login,
		Password: password,
		Email:    email,
	}

	testCore.CreateUserAccount(testRequest)

	_, foundAccount := testCore.FindUserAccount(login)
	if !foundAccount {
		t.Errorf("user not found")
	}
}

func TestCreateAndKillSession(t *testing.T) {
	login := "testLogin"
	testCore := Core{sessions: make(map[string]Session)}

	sid, _ := testCore.CreateSession(login)
	isFound := testCore.FindActiveSession(sid)
	if !isFound {
		t.Errorf("session not found")
	}

	testCore.KillSession(sid)

	isFound = testCore.FindActiveSession(sid)
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
	json.Unmarshal(body, &response)

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
	json.Unmarshal(body, &response)

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
	json.Unmarshal(body, &response)

	if response.Status != http.StatusMethodNotAllowed {
		t.Errorf("got incorrect status")
	}
}

func TestFilmsPages(t *testing.T) {
	h1 := httptest.NewRequest(http.MethodGet, "/api/v1/films?page=100", nil)
	h2 := httptest.NewRequest(http.MethodGet, "/api/v1/films?page=2", nil)
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	api := API{}
	api.Films(w1, h1)
	api.Films(w2, h2)

	var response1, response2 Response

	body1, _ := io.ReadAll(w1.Body)
	json.Unmarshal(body1, &response1)
	body2, _ := io.ReadAll(w2.Body)
	json.Unmarshal(body2, &response2)

	if !reflect.DeepEqual(response1.Body, response2.Body) {
		t.Errorf("pages not matching")
	}
}
