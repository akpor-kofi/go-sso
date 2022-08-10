package ports

import "go-sso/internal/core/domain"

type UserService interface {
	GetAll() ([]*domain.User, error)
	Get(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	New(user *domain.User) (*domain.User, error)
	Update(id string, user *domain.User) (*domain.User, error)
	Delete(id string) error
	UpdateResetToken(email string, resetToken string) error
}

type CompanyService interface {
	GetAll() ([]*domain.Company, error)
	Get(id string) (*domain.Company, error)
	New(company *domain.Company, owner *domain.User) (*domain.Company, error)
	Update(id string, company *domain.Company) (*domain.Company, error)
	Delete(id string) error
	GetCompanyRole(companyId, userId string) string
	AddEmployee(companyId, userId, role string) error
}

type ClientAppService interface {
	New(cappName string, owner *domain.User) (*domain.ClientApp, error)
	GetAll(opts ...string) ([]*domain.ClientApp, error)
	Get(clientId string) (*domain.ClientApp, error)
	Delete(clientId string) error
	AuthorizeClientCredentials(requestToken, clientId string) (*domain.ClientApp, error)
}
