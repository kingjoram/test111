package crew

type CrewItem struct {
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	Birthdate string `json:"birth_date"`
	Photo     string `json:"photo"`
}

type Character struct {
	IdActor       uint64 `json:"id_actor"`
	ActorPhoto    string `json:"photo_actor"`
	NameActor     string `json:"name_actor"`
	NameCharacter string `json:"name_character"`
}
