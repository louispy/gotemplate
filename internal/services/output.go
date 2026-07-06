package services

type UserOutput struct {
	Id        string
	Name      string
	Email     string
	CreatedAt string
	UpdatedAt string
}

type ListUsersOutput struct {
	Users []UserOutput
}
