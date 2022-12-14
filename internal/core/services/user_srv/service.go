package user_srv

import (
	"go-sso/internal/core/domain"
	"go-sso/internal/core/ports"

	"github.com/gofiber/fiber/v2/utils"
)

type service struct {
	userRepository ports.UserRepository
}

func New(userRepository ports.UserRepository) *service {
	return &service{
		userRepository,
	}
}

func (s *service) GetAll() ([]*domain.User, error) {
	return s.userRepository.GetAll()
}

func (s *service) Get(id string) (*domain.User, error) {
	user, err := s.userRepository.Get(id)
	user.Post("find")
	return user, err
}

func (s *service) GetByEmail(email string) (*domain.User, error) {
	return s.userRepository.GetByEmail(email)
}

func (s *service) New(user *domain.User) (*domain.User, error) {
	user.Id = utils.UUIDv4()
	user.Pre("save", "new")
	newUser, err := s.userRepository.New(user)
	newUser.Post("find")
	return newUser, err

}

func (s *service) Update(id string, user *domain.User) (*domain.User, error) {
	return s.userRepository.Update(id, user)
}

func (s *service) Delete(id string) error {
	return s.userRepository.Delete(id)
}

func (s *service) UpdateResetToken(email string, resetToken string) error {
	return s.userRepository.UpdateResetToken(email, resetToken)
}

func (s *service) GetResetToken(token string) (*domain.User, error) {
	return s.userRepository.GetResetToken(token)
}

func (s *service) UpdatePassword(id, password string) error {
	return s.userRepository.UpdatePassword(id, password)
}
