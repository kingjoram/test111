package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
