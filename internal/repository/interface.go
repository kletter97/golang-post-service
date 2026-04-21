package repository


type TestRepository interface {
	GetGreeting() (string, error)
}
