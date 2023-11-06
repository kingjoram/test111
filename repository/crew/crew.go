package crew

type CrewItem struct {
	Id        int
	Name      string
	Birthdate string
	Photo     string
}

type Character struct {
	IdActor       uint64
	ActorPhoto    string
	NameActor     string
	NameCharacter string
}
