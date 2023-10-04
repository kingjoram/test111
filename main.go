package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {

	logFile, _ := os.Create("log.log")
	lg := slog.New(slog.NewJSONHandler(logFile, nil))

	core := Core{
		sessions: make(map[string]Session),
		users:    make(map[string]User),
		collections: map[string]string{
			"new":       "Новинки",
			"action":    "Боевик",
			"comedy":    "Комедия",
			"ru":        "Российский",
			"eu":        "Зарубежный",
			"war":       "Военный",
			"kids":      "Детский",
			"detective": "Детектив",
			"drama":     "Драма",
			"crime":     "Криминал",
			"melodrama": "Мелодрама",
			"horror":    "Ужас",
		},
		lg: lg.With("module", "core"),
	}
	api := API{core: &core, lg: lg.With("module", "api")}

	mx := http.NewServeMux()
	mx.HandleFunc("/signup", api.Signup)
	mx.HandleFunc("/signin", api.Signin)
	mx.HandleFunc("/logout", api.LogoutSession)
	mx.HandleFunc("/authcheck", api.AuthAccept)
	mx.HandleFunc("/api/v1/films", api.Films)
	http.ListenAndServe(":8080", mx)
}
