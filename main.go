package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/comment"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/crew"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/film"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/genre"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/profile"
)

func main() {
	logFile, _ := os.Create("log.log")
	lg := slog.New(slog.NewJSONHandler(logFile, nil))

	config, err := configs.ReadConfig()
	if err != nil {
		lg.Error("read config error", "err", err.Error())
		return
	}
	core := Core{
		sessions: make(map[string]Session),
		lg:       lg.With("module", "core"),
		Films:    film.GetFilmRepo(*config, lg),
		Users:    profile.GetUserRepo(*config, lg),
		Genres:   genre.GetGenreRepo(*config, lg),
		Comments: comment.GetCommentRepo(*config, lg),
		Crew:     crew.GetCrewRepo(*config, lg),
	}
	api := API{core: &core, lg: lg.With("module", "api")}

	go core.PingRepos(config.Timer)

	mx := http.NewServeMux()
	mx.HandleFunc("/signup", api.Signup)
	mx.HandleFunc("/signin", api.Signin)
	mx.HandleFunc("/logout", api.LogoutSession)
	mx.HandleFunc("/authcheck", api.AuthAccept)
	mx.HandleFunc("/api/v1/films", api.Films)
	mx.HandleFunc("/api/v1/film", api.Film)
	err = http.ListenAndServe(":8080", mx)
	if err != nil {
		api.lg.Error("ListenAndServe error", "err", err.Error())
	}
}
