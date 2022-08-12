package client_app_srv

import (
	"go-sso/internal/core/domain"
	"go-sso/internal/core/ports"
)

type service struct {
	clientAppRepository ports.ClientAppRepository
}

func New(clientAppRepository ports.ClientAppRepository) *service {
	return &service{clientAppRepository}
}

func (s *service) New(appName string, owner *domain.User) (*domain.ClientApp, error) {
	clientApp := domain.New(appName)
	return s.clientAppRepository.New(clientApp, owner)
}

func (s *service) GetAll(opts ...string) ([]*domain.ClientApp, error) {
	return s.clientAppRepository.GetAll(opts...)
}

func (s *service) Get(clientId string) (*domain.ClientApp, error) {
	return s.clientAppRepository.Get(clientId)
}

func (s *service) Delete(id string) error {
	return s.clientAppRepository.Delete(id)
}

func (s *service) AuthorizeClientCredentials(requestToken, clientId string) (*domain.ClientApp, error) {
	return s.clientAppRepository.AuthorizeClientCredentials(requestToken, clientId)
}
