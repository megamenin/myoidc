package domain

type User struct {
	Id          string   `json:"id"`
	Login       string   `json:"login"`
	Email       string   `json:"email"`
	FullName    string   `json:"fullName"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Permissions []string `json:"permissions"`
}
