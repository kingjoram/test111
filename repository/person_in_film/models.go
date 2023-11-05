package person_in_film_repo

import "database/sql"

type PersonInFilmItem struct {
	IdFilm        int
	IdPerson      int
	IdProfession  int
	CharacterName sql.NullString
}
