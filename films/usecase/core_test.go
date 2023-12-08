package usecase

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/films/mocks"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/requests"
	"github.com/golang/mock/gomock"
)

func TestGetCalendar(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedDay := models.DayItem{DayNumber: 1, DayNews: "n"}
	expectedDays := []models.DayItem{expectedDay}
	expected := &requests.CalendarResponse{MonthName: time.Now().Month().String(), MonthText: "Новинки этого месяца", CurrentDay: uint8(time.Now().Day()), Days: expectedDays}

	mockObj := mocks.NewMockICalendarRepo(mockCtrl)
	firstCall := mockObj.EXPECT().GetCalendar().Return(expectedDays, nil)
	mockObj.EXPECT().GetCalendar().After(firstCall).Return(nil, fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{calendar: mockObj, lg: logger}

	result, err := core.GetCalendar()
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.GetCalendar()
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}
}

func TestGetActorsCareer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedProf := models.ProfessionItem{Id: 1, Title: "p1"}
	expected := []models.ProfessionItem{expectedProf}

	mockObj := mocks.NewMockIProfessionRepo(mockCtrl)
	firstCall := mockObj.EXPECT().GetActorsProfessions(uint64(1)).Return(expected, nil)
	mockObj.EXPECT().GetActorsProfessions(uint64(1)).After(firstCall).Return(nil, fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{profession: mockObj, lg: logger}

	result, err := core.GetActorsCareer(1)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.GetActorsCareer(1)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}
}

func TestGetGenre(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expected := "g1"

	mockObj := mocks.NewMockIGenreRepo(mockCtrl)
	firstCall := mockObj.EXPECT().GetGenreById(uint64(1)).Return(expected, nil)
	mockObj.EXPECT().GetGenreById(uint64(1)).After(firstCall).Return("", fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{genres: mockObj, lg: logger}

	result, err := core.GetGenre(1)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.GetGenre(1)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != "" {
		t.Errorf("unexpected result")
		return
	}
}

func TestFindFilm(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedFilm := models.FilmItem{Title: "t"}
	expected := []models.FilmItem{expectedFilm}

	mockObj := mocks.NewMockIFilmsRepo(mockCtrl)
	firstCall := mockObj.EXPECT().FindFilm(string("t"), string("df"), string("dt"), float32(0), float32(10), string(""), nil, nil).Return(expected, nil)
	mockObj.EXPECT().FindFilm(string("t0"), string("df"), string("dt"), float32(0), float32(10), string(""), nil, nil).After(firstCall).Return(nil, fmt.Errorf("repo_error"))
	mockObj.EXPECT().FindFilm(string("t10"), string("df"), string("dt"), float32(0), float32(10), string(""), nil, nil).Return([]models.FilmItem{}, nil)

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{films: mockObj, lg: logger}

	result, err := core.FindFilm("t", "df", "dt", 0, 10, "", nil, nil)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.FindFilm("t0", "df", "dt", 0, 10, "", nil, nil)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}

	result, err = core.FindFilm("t10", "df", "dt", 0, 10, "", nil, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected not found")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}
}

func TestFindActor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedFilm := models.Character{NameActor: "t"}
	expected := []models.Character{expectedFilm}

	mockObj := mocks.NewMockICrewRepo(mockCtrl)
	firstCall := mockObj.EXPECT().FindActor(string("t"), string("bd"), nil, nil, string("")).Return(expected, nil)
	mockObj.EXPECT().FindActor(string("t"), string("bd"), nil, nil, string("")).After(firstCall).Return(nil, fmt.Errorf("repo_error"))
	mockObj.EXPECT().FindActor(string("t"), string("bd"), nil, nil, string("")).Return([]models.Character{}, nil)

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{crew: mockObj, lg: logger}

	result, err := core.FindActor("t", "bd", nil, nil, "")
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.FindActor("t", "bd", nil, nil, "")
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}

	result, err = core.FindActor("t", "bd", nil, nil, "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected not found")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}
}

func TestGetFilmsAndGenreTitle(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedFilm := models.FilmItem{Title: "t"}
	expectedFilms := []models.FilmItem{expectedFilm}

	expectedGenre := "g1"

	mockObj := mocks.NewMockIFilmsRepo(mockCtrl)
	mockObj.EXPECT().GetFilms(uint64(1), uint64(1)).Return(expectedFilms, nil)
	mockObj.EXPECT().GetFilms(uint64(1), uint64(0)).Return(nil, fmt.Errorf("repo_error"))
	mockObj.EXPECT().GetFilmsByGenre(uint64(10), uint64(1), uint64(1)).Return(expectedFilms, nil)

	mockGenres := mocks.NewMockIGenreRepo(mockCtrl)
	mockGenres.EXPECT().GetGenreById(uint64(0)).Return(expectedGenre, nil)
	mockGenres.EXPECT().GetGenreById(uint64(10)).Return("", fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{films: mockObj, genres: mockGenres, lg: logger}

	films, genre, err := core.GetFilmsAndGenreTitle(0, 1, 1)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expectedFilms, films) {
		t.Errorf("wanted %v, had %v", expectedFilms, films)
		return
	}
	if genre != expectedGenre {
		t.Errorf("wanted %v, had %v", expectedGenre, genre)
		return
	}

	films, genre, err = core.GetFilmsAndGenreTitle(0, 1, 0)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if films != nil {
		t.Errorf("unexpected result")
		return
	}
	if genre != "" {
		t.Errorf("unexpected result")
		return
	}

	films, genre, err = core.GetFilmsAndGenreTitle(10, 1, 1)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if films != nil {
		t.Errorf("unexpected result")
		return
	}
	if genre != "" {
		t.Errorf("unexpected result")
		return
	}
}

func TestGetActorInfo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expProf := models.ProfessionItem{Title: "t"}
	expectedCareer := []models.ProfessionItem{expProf}
	expectedActor := &models.CrewItem{Name: "n"}
	expected := &requests.ActorResponse{Name: expectedActor.Name, Career: expectedCareer}

	mockObj := mocks.NewMockICrewRepo(mockCtrl)
	firstCall := mockObj.EXPECT().GetActor(uint64(1)).Return(expectedActor, nil)
	mockObj.EXPECT().GetActor(uint64(2)).After(firstCall).Return(nil, fmt.Errorf("repo_error"))
	mockObj.EXPECT().GetActor(uint64(3)).Return(&models.CrewItem{}, nil)
	mockObj.EXPECT().GetActor(uint64(4)).Return(expectedActor, nil)

	mockProf := mocks.NewMockIProfessionRepo(mockCtrl)
	mockProf.EXPECT().GetActorsProfessions(uint64(1)).Return(expectedCareer, nil)
	mockProf.EXPECT().GetActorsProfessions(uint64(4)).Return(nil, fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{crew: mockObj, profession: mockProf, lg: logger}

	result, err := core.GetActorInfo(1)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.GetActorInfo(2)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}

	result, err = core.GetActorInfo(3)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected not found")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}

	result, err = core.GetActorInfo(4)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}
}

func TestGetFilmInfo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedFilm := &models.FilmItem{Title: "t"}

	genreItem := models.GenreItem{Title: "g1"}
	expectedGenres := []models.GenreItem{genreItem}

	crewItem := models.CrewItem{Name: "n"}
	expectedCrew := []models.CrewItem{crewItem}

	charItem := models.Character{NameActor: "an"}
	expectedCharacters := []models.Character{charItem}
	expectedRating := 9.8
	expectedNumber := uint64(100)
	expectedResult := &requests.FilmResponse{
		Film:       *expectedFilm,
		Genres:     expectedGenres,
		Directors:  expectedCrew,
		Scenarists: expectedCrew,
		Characters: expectedCharacters,
		Rating:     expectedRating,
		Number:     expectedNumber}

	mockFilm := mocks.NewMockIFilmsRepo(mockCtrl)
	notFound := mockFilm.EXPECT().GetFilm(uint64(1)).Return(&models.FilmItem{}, nil).Times(1)
	withErr := mockFilm.EXPECT().GetFilm(uint64(1)).Return(nil, fmt.Errorf("repo_error")).Times(1).After(notFound)
	mockFilm.EXPECT().GetFilm(uint64(1)).Return(expectedFilm, nil).After(withErr).AnyTimes()

	mockGenres := mocks.NewMockIGenreRepo(mockCtrl)
	withErr = mockGenres.EXPECT().GetFilmGenres(uint64(1)).Return(nil, fmt.Errorf("repo_error")).Times(1)
	mockGenres.EXPECT().GetFilmGenres(uint64(1)).Return(expectedGenres, nil).AnyTimes().After(withErr)

	withErr = mockFilm.EXPECT().GetFilmRating(uint64(1)).Return(float64(0), uint64(0), fmt.Errorf("repo_error")).Times(1)
	mockFilm.EXPECT().GetFilmRating(uint64(1)).Return(expectedRating, expectedNumber, nil).AnyTimes().After(withErr)

	mockCrew := mocks.NewMockICrewRepo(mockCtrl)
	withErr = mockCrew.EXPECT().GetFilmDirectors(uint64(1)).Return(nil, fmt.Errorf("repo_error")).Times(1)
	mockCrew.EXPECT().GetFilmDirectors(uint64(1)).Return(expectedCrew, nil).AnyTimes().After(withErr)

	withErr = mockCrew.EXPECT().GetFilmScenarists(uint64(1)).Return(nil, fmt.Errorf("repo_error")).Times(1)
	mockCrew.EXPECT().GetFilmScenarists(uint64(1)).Return(expectedCrew, nil).AnyTimes().After(withErr)

	withErr = mockCrew.EXPECT().GetFilmCharacters(uint64(1)).Return(nil, fmt.Errorf("repo_error")).Times(1)
	mockCrew.EXPECT().GetFilmCharacters(uint64(1)).Return(expectedCharacters, nil).AnyTimes().After(withErr)

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{films: mockFilm, genres: mockGenres, crew: mockCrew, lg: logger}

	result, err := core.GetFilmInfo(1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("wanted not found error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}

	for i := 0; i < 6; i++ {
		result, err = core.GetFilmInfo(1)
		if err == nil {
			t.Errorf("wanted error")
			return
		}
		if result != nil {
			t.Errorf("unexpected result")
			return
		}
	}

	result, err = core.GetFilmInfo(1)
	if err != nil {
		t.Errorf("wanted no errors")
		return
	}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("unexpected result. wanted %v, got %v", expectedResult, result)
		return
	}
}

func TestFavoriteFilms(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedFilm := models.FilmItem{Title: "t"}
	expected := []models.FilmItem{expectedFilm}

	mockObj := mocks.NewMockIFilmsRepo(mockCtrl)
	firstCall := mockObj.EXPECT().GetFavoriteFilms(uint64(1), uint64(1), uint64(1)).Return(expected, nil)
	mockObj.EXPECT().GetFavoriteFilms(uint64(1), uint64(1), uint64(1)).After(firstCall).Return(nil, fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{films: mockObj, lg: logger}

	result, err := core.FavoriteFilms(1, 1, 1)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("wanted %v, had %v", expected, result)
		return
	}

	result, err = core.FavoriteFilms(1, 1, 1)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
	if result != nil {
		t.Errorf("unexpected result")
		return
	}
}

func TestFavoriteFilmsRemove(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := mocks.NewMockIFilmsRepo(mockCtrl)
	firstCall := mockObj.EXPECT().RemoveFavoriteFilm(uint64(1), uint64(1)).Return(nil)
	mockObj.EXPECT().RemoveFavoriteFilm(uint64(1), uint64(1)).After(firstCall).Return(fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{films: mockObj, lg: logger}

	err := core.FavoriteFilmsRemove(1, 1)
	if err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}

	err = core.FavoriteFilmsRemove(1, 1)
	if err == nil {
		t.Errorf("wanted error")
		return
	}
}

func TestFavoriteFilmsAdd(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := mocks.NewMockIFilmsRepo(mockCtrl)
	mockObj.EXPECT().CheckFilm(uint64(1), uint64(1)).Return(true, nil).Times(1).Times(1)
	mockObj.EXPECT().CheckFilm(uint64(1), uint64(1)).Return(false, fmt.Errorf("repo_err")).Times(1)
	mockObj.EXPECT().CheckFilm(uint64(1), uint64(1)).Return(false, nil).Times(2)

	mockObj.EXPECT().AddFavoriteFilm(uint64(1), uint64(1)).Return(fmt.Errorf("repo_error")).Times(1)
	mockObj.EXPECT().AddFavoriteFilm(uint64(1), uint64(1)).Return(nil).Times(1)

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{films: mockObj, lg: logger}

	err := core.FavoriteFilmsAdd(1, 1)
	if !errors.Is(err, ErrFoundFavorite) {
		t.Errorf("expected found error, got %s", err)
		return
	}

	for i := 0; i < 2; i++ {
		err = core.FavoriteFilmsAdd(1, 1)
		if err == nil {
			t.Errorf("wanted error")
			return
		}
	}

	err = core.FavoriteFilmsAdd(1, 1)
	if err != nil {
		t.Errorf("unexpected error")
		return
	}
}
