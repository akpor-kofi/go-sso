package company_srv

import (
	"go-sso/internal/core/domain"
	"go-sso/internal/core/ports"

	"github.com/gofiber/fiber/v2/utils"
)

type service struct {
	companyRepository ports.CompanyRepository
}

func New(companyRepository ports.CompanyRepository) *service {
	return &service{companyRepository}
}

func (s *service) GetAll() ([]*domain.Company, error) {
	// add filters
	return s.companyRepository.GetAll()
}

func (s *service) Get(id string) (*domain.Company, error) {
	return s.companyRepository.Get(id)
}

func (s *service) New(company *domain.Company, owner *domain.User) (*domain.Company, error) {
	company.Id = utils.UUIDv4()
	company.Pre("save")
	return s.companyRepository.New(company, owner)
}

func (s *service) Update(id string, company *domain.Company) (*domain.Company, error) {
	return s.companyRepository.Update(id, company)
}

func (s *service) Delete(id string) error {
	return s.companyRepository.Delete(id)
}

func (s *service) GetCompanyRole(companyId, userId string) string {
	return s.companyRepository.GetCompanyRole(companyId, userId)
}

func (s *service) AddEmployee(companyId, userId, role string) error {
	return s.companyRepository.AddEmployee(companyId, userId, role)
}
