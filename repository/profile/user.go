package profile

type UserItem struct {
	Id               uint64 `sql:"AUTO_INCREMENT"`
	Name             string
	Birthdate        string
	Photo            string
	Login            string
	Password         string
	RegistrationDate string
	Email            string
}
