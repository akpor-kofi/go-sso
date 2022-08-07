package ports

import "go-sso/internal/core/domain"

// also add a filter query in the getAll

type UserRepository interface {
	GetAll() ([]*domain.User, error)
	Get(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	New(user *domain.User) (*domain.User, error)
	Update(id string, user *domain.User) (*domain.User, error)
	Delete(id string) error
}

type CompanyRepository interface {
	GetAll() ([]*domain.Company, error)
	Get(id string) (*domain.Company, error)
	New(company *domain.Company, owner *domain.User) (*domain.Company, error)
	Update(id string, company *domain.Company) (*domain.Company, error)
	Delete(id string) error
	GetCompanyRole(companyId, userId string) string
	AddEmployee(companyId, userId, role string) error
}

type ClientAppRepository interface {
	New(clientApp *domain.ClientApp, owner *domain.User) (*domain.ClientApp, error)
	GetAll(opts ...string) ([]*domain.ClientApp, error)
	Get(clientId string) (*domain.ClientApp, error)
	Delete(clientId string) error
	AuthorizeClientCredentials(requestToken, clientId string) (*domain.ClientApp, error)
}
