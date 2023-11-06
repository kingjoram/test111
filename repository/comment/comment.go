package comment

type CommentItem struct {
	Username string `json:"login"`
	IdFilm   uint64 `json:"id_film"`
	Rating   uint64 `json:"rating"`
	Comment  string `json:"comment"`
}
