package neo4j

import (
	"fmt"
	"go-sso/internal/core/domain"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

func deserializeUsers(usersMap []map[string]interface{}) []*domain.User {
	users := new([]*domain.User)

	err := mapstructure.Decode(usersMap, users)
	if err != nil {
		return nil
	}

	return *users
}

func deserializeCompanies(companiesMap []map[string]interface{}) []*domain.Company {
	companies := new([]*domain.Company)

	err := mapstructure.Decode(companiesMap, companies)
	if err != nil {
		return nil
	}

	return *companies
}

func deserializeUser(userMap map[string]interface{}) (*domain.User, error) {
	user := new(domain.User)

	err := mapstructure.Decode(userMap, user)
	if err != nil {
		return &domain.User{}, err
	}

	if structs.IsZero(user) {
		return &domain.User{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func deserializeCompany(companyMap map[string]interface{}) (*domain.Company, error) {
	company := new(domain.Company)

	err := mapstructure.Decode(companyMap, company)
	if err != nil {
		return &domain.Company{}, err
	}

	if structs.IsZero(company) {
		return &domain.Company{}, fmt.Errorf("company not found")
	}

	return company, nil
}

func deserializeClientApp(clientAppMap map[string]interface{}) (*domain.ClientApp, error) {
	clientApp := new(domain.ClientApp)

	err := mapstructure.Decode(clientAppMap, clientApp)
	if err != nil {
		return &domain.ClientApp{}, err
	}

	if structs.IsZero(clientApp) {
		return &domain.ClientApp{}, fmt.Errorf("Client Application not found")
	}

	return clientApp, nil
}

func serializeDataToMap(o interface{}) map[string]interface{} {
	return structs.Map(o)
}
