package services

import (
	//    "github.com/google/uuid"
	"github.com/mezni/expenses-go/internal/domain/entities"
	"github.com/mezni/expenses-go/internal/domain/repositories"
)

type ExpenseService struct {
	repo repositories.OrgRepository
}

func NewExpenseService(repo repositories.OrgRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) FindByName(name string) (*entities.Org, error) {
	return s.repo.FindByName(name)
}
