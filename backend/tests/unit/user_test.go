package unit

import (
	"mobile-backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestUserModel_HashPassword(t *testing.T) {
	user := &models.User{}
	password := "testpassword123"

	err := user.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, user.Password)
	assert.NotEqual(t, password, user.Password)
}

func TestUserModel_CheckPassword(t *testing.T) {
	user := &models.User{}
	password := "testpassword123"

	// Hash the password first
	err := user.HashPassword(password)
	assert.NoError(t, err)

	// Check correct password
	err = user.CheckPassword(password)
	assert.NoError(t, err)

	// Check incorrect password
	err = user.CheckPassword("wrongpassword")
	assert.Error(t, err)
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	// userService := NewUserService(mockRepo)

	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
	}

	mockRepo.On("Create", user).Return(nil)

	// Test would go here with actual service implementation
	// err := userService.CreateUser(user)
	// assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)

	email := "test@example.com"
	expectedUser := &models.User{
		Email: email,
		Name:  "Test User",
	}

	mockRepo.On("GetByEmail", email).Return(expectedUser, nil)

	// Test would go here with actual service implementation
	// user, err := userService.GetUserByEmail(email)
	// assert.NoError(t, err)
	// assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}
