package requests

//easyjson:json
type (
	SignupRequest struct {
		Login     string `json:"login"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		BirthDate string `json:"birth_date"`
		Name      string `json:"name"`
	}

	SigninRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	CommentRequest struct {
		FilmId uint64 `json:"film_id"`
		Rating uint16 `json:"rating"`
		Text   string `json:"text"`
	}

	EditProfileRequest struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Photo    []byte `json:"photo"`
		Password string `json:"password"`
	}

	FindFilmRequest struct {
		Title      string   `json:"title"`
		DateFrom   string   `json:"date_from"`
		DateTo     string   `json:"date_to"`
		RatingFrom float32  `json:"rating_from"`
		RatingTo   float32  `json:"rating_to"`
		Mpaa       string   `json:"mpaa"`
		Genres     []uint32 `json:"genres"`
		Actors     []string `json:"actors"`
	}

	FindActorRequest struct {
		Name      string   `json:"name"`
		BirthDate string   `json:"birthday"`
		Career    []string `json:"amplua"`
		Films     []string `json:"films"`
		Country   string   `json:"country"`
	}
)
