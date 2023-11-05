package film

type FilmItem struct {
	Id          uint32 `sql:"AUTO_INCREMENT"`
	Title       string
	Info        string
	Poster      string
	ReleaseDate string
	Country     string
	Mpaa        string
}
