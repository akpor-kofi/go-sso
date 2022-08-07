package neo4j

import (
	"go-sso/internal/core/domain"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type CompanyStorage struct {
}

func NewCompanyStorage() *CompanyStorage {
	return &CompanyStorage{}
}

func (c CompanyStorage) GetAll() ([]*domain.Company, error) {
	result, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (c:Company)
				RETURN c
   		`

		result, err := tx.Run(cypher, map[string]interface{}{})

		if err != nil {
			return nil, err
		}

		companies := make([]map[string]interface{}, 0) // len doesn't really matter

		for result.Next() {
			company := result.Record().Values[0].(neo4j.Node).Props

			companies = append(companies, company)

		}

		return deserializeCompanies(companies), result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*domain.Company), nil

}

func (c CompanyStorage) Get(id string) (*domain.Company, error) {
	records, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (c:Company)
				WHERE c.id = $companyId
				RETURN c
   		`

		result, err := tx.Run(cypher, map[string]interface{}{"companyId": id})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	var companyMap map[string]interface{}

	for _, record := range records.([]*neo4j.Record) {
		companyMap = record.Values[0].(neo4j.Node).Props
	}

	company, err := deserializeCompany(companyMap)

	if err != nil {
		return &domain.Company{}, err
	}

	return company, nil
}

func (c CompanyStorage) New(company *domain.Company, owner *domain.User) (*domain.Company, error) {
	properties := serializeDataToMap(company)
	locationProps := serializeDataToMap(company.Location)

	_, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		createCompanyQuery := `
			CREATE (c:Company {id: $Id, name: $Name, cacNum: $CacNum, email: $Email, phoneNumber: $PhoneNumber, currency: $Currency, image: $Image, createdAt: $CreatedAt, updatedAt: $UpdatedAt, version: $Version})
			`
		createLocationQuery := `CREATE (l:Location {lng: $Lng, lat: $Lat, address: $Address, description: $Description})`
		createOwnerCompanyRelationshipQuery := `
			MERGE (u:User { id: $userId })
			MERGE (c:Company { id: $companyId })
			MERGE (l:Location { lng: $lng, lat: $lat})
			MERGE (u)-[o: OWNS { CreatedAt: $timestamp}]->(c)
			MERGE (u)-[w: WORKS_IN { CreatedAt: $timestamp, role: $role}]->(c)
			MERGE (c)-[a: LOCATED_AT]->(l)
			RETURN u, c, o
			`

		tx.Run(createCompanyQuery, properties)
		tx.Run(createLocationQuery, locationProps)
		result, err := tx.Run(createOwnerCompanyRelationshipQuery, map[string]interface{}{
			"userId":    owner.Id,
			"companyId": company.Id,
			"timestamp": time.Now().UnixMilli(),
			"role":      "owner",
			"lng":       company.Lng,
			"lat":       company.Lat,
		})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	return company, nil
}

func (c CompanyStorage) Update(id string, company *domain.Company) (*domain.Company, error) {
	//TODO implement me
	panic("implement me")
}

func (c CompanyStorage) Delete(id string) error {
	_, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (c:Company)
				WHERE c.id = $companyId
				DELETE c
   		`

		result, err := tx.Run(cypher, map[string]interface{}{"companyId": id})
		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		return err
	}

	return nil
}

func (c CompanyStorage) GetCompanyRole(companyId, userId string) string {
	records, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		getUserCompanyRole := `
			MATCH (u:User {id:$userId})-[w:WORKS_IN]->(c:Company {id:$companyId})
			RETURN w
		`
		result, err := tx.Run(getUserCompanyRole, map[string]interface{}{"userId": userId, "companyId": companyId})
		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	var rel string

	for _, record := range records.([]*neo4j.Record) {
		rel = record.Values[0].(neo4j.Relationship).Props["role"].(string)
	}

	return rel
}

func (c CompanyStorage) AddEmployee(companyId, userId, role string) error {
	_, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		addEmployeeQuery := `
			MERGE (u:User { id: $userId })
			MERGE (c:Company { id: $companyId })
			MERGE (u)-[w: WORKS_IN { CreatedAt: $timestamp, role: $role}]->(c)
		`
		result, err := tx.Run(addEmployeeQuery, map[string]interface{}{
			"userId":    userId,
			"companyId": companyId,
			"timestamp": time.Now().UnixMilli(),
			"role":      role,
		})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		return err
	}

	return nil
}
