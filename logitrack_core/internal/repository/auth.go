package repository

import (
	"fmt"
	"sync"

	"github.com/logitrack/core/internal/model"
)

type AuthRepository interface {
	FindUser(username, password string) (model.User, error)
	SaveToken(token string, user model.User)
	GetUserByToken(token string) (model.User, error)
	DeleteToken(token string)
	ListByRole(role model.Role) []model.User
	GetUserByID(id string) (model.User, error)
}

type credential struct {
	user     model.User
	password string
}

type inMemoryAuthRepository struct {
	mu     sync.RWMutex
	tokens map[string]model.User
	users  []credential
}

func NewInMemoryAuthRepository() AuthRepository {
	return &inMemoryAuthRepository{
		tokens: make(map[string]model.User),
		users: []credential{
			{user: model.User{ID: "1", Username: "operator", Role: model.RoleOperator}, password: "operator123"},
			{user: model.User{ID: "2", Username: "supervisor", Role: model.RoleSupervisor}, password: "supervisor123"},
			{user: model.User{ID: "3", Username: "gerente", Role: model.RoleManager}, password: "gerente123"},
			{user: model.User{ID: "4", Username: "admin", Role: model.RoleAdmin}, password: "admin123"},
			{user: model.User{ID: "5", Username: "chofer", Role: model.RoleDriver}, password: "chofer123"},
		},
	}
}

func (r *inMemoryAuthRepository) FindUser(username, password string) (model.User, error) {
	for _, c := range r.users {
		if c.user.Username == username && c.password == password {
			return c.user, nil
		}
	}
	return model.User{}, fmt.Errorf("invalid credentials")
}

func (r *inMemoryAuthRepository) SaveToken(token string, user model.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = user
}

func (r *inMemoryAuthRepository) GetUserByToken(token string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.tokens[token]
	if !ok {
		return model.User{}, fmt.Errorf("invalid token")
	}
	return user, nil
}

func (r *inMemoryAuthRepository) DeleteToken(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, token)
}

func (r *inMemoryAuthRepository) ListByRole(role model.Role) []model.User {
	var result []model.User
	for _, c := range r.users {
		if c.user.Role == role {
			result = append(result, c.user)
		}
	}
	return result
}

func (r *inMemoryAuthRepository) GetUserByID(id string) (model.User, error) {
	for _, c := range r.users {
		if c.user.ID == id {
			return c.user, nil
		}
	}
	return model.User{}, fmt.Errorf("user not found")
}
