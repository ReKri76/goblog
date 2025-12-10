package tests

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

// Интерфейсы объявляем первыми
type ValidationService interface {
	ValidateUsername(username string) error
	ValidatePassword(password string) error
}

type UserRepository interface {
	Authenticate(username string, password string) (bool, error)
}

type AuthService interface {
	Authenticate(username string, password string) (bool, error)
}

// Реализации объявляем после интерфейсов
type ValidationServiceImpl struct{}

func (svc *ValidationServiceImpl) ValidateUsername(username string) error {
	// perform validation logic
	return nil // return nil if validation succeeds, or an error if it fails
}

func (svc *ValidationServiceImpl) ValidatePassword(password string) error {
	// perform validation logic
	return nil // return nil if validation succeeds, or an error if it fails
}

type UserRepositoryImpl struct{}

func (repo *UserRepositoryImpl) Authenticate(username string, password string) (bool, error) {
	// perform authentication logic
	return true, nil
}

type AuthServiceImpl struct {
	validator ValidationService
	repo      UserRepository
}

func NewAuthServiceImpl(validator ValidationService, repo UserRepository) AuthService {
	return &AuthServiceImpl{
		validator: validator,
		repo:      repo,
	}
}

func (s *AuthServiceImpl) Authenticate(username string, password string) (bool, error) {
	if err := s.validator.ValidateUsername(username); err != nil {
		return false, err
	}
	if err := s.validator.ValidatePassword(password); err != nil {
		return false, err
	}
	return s.repo.Authenticate(username, password)
}

type MockValidationService struct {
	mock.Mock
}

func (m *MockValidationService) ValidateUsername(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockValidationService) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Authenticate(username string, password string) (bool, error) {

	args := m.Called(username, password)
	return args.Bool(0), args.Error(1)
}

func TestAuthenticate(t *testing.T) {
	username := "testuser"
	password := "testpassword"

	mockValidator := new(MockValidationService)
	mockValidator.On("ValidateUsername", username).Return(nil)
	mockValidator.On("ValidatePassword", password).Return(nil)
	mockRepo := new(MockUserRepository)
	mockRepo.On("Authenticate", username, password).Return(true, nil)
	authService := NewAuthServiceImpl(mockValidator, mockRepo)
	authenticated, err := authService.Authenticate(username, password)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !authenticated {
		t.Error("Expected authentication to succeed, but it failed")
	}
}
