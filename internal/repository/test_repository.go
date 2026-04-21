package repository

import "fmt"

type testRepository struct{}

func NewTestRepository() TestRepository {
	return &testRepository{}
}

func (r *testRepository) GetGreeting() (string, error) {
	greeting := "Hello!"
	return greeting, nil
}