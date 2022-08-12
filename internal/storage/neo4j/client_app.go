package neo4j

import (
	"go-sso/internal/core/domain"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type ClientAppStorage struct {
}

func NewClientAppStorage() *ClientAppStorage {
	return &ClientAppStorage{}
}

// type ClientAppRepository interface {
// 	New(clientApp *domain.ClientApp, owner *domain.User) (*domain.ClientApp, error)
// 	GetAll(opts ...string) ([]*domain.ClientApp, error)
// 	Get(clientId string) (*domain.ClientApp, error)
// 	Delete(clientId string) error
// }

func (c ClientAppStorage) New(clientApp *domain.ClientApp, owner *domain.User) (*domain.ClientApp, error) {
	properties := serializeDataToMap(clientApp)
	records, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		createClientApplicationQuery := `
			CREATE (a:ClientApplication { id: $Id, appName: $AppName, requestToken: $RequestToken, secret: $Secret } )
		`

		createRelationshipQuery := `
			MERGE (u:User { id: $userId })
			MERGE (a:ClientApplication { id: $clientId })
			MERGE (u)-[c: CREATED {CreatedAt: $timestamp}]->(a)
			RETURN a
		`

		_, err := tx.Run(createClientApplicationQuery, properties)
		result, err := tx.Run(createRelationshipQuery, map[string]interface{}{
			"userId":    owner.Id,
			"clientId":  clientApp.Id,
			"timestamp": time.Now().UnixMilli(),
		})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		return nil, err
	}

	var appDetailsMap map[string]interface{}
	for _, record := range records.([]*neo4j.Record) {
		appDetailsMap = record.Values[0].(neo4j.Node).Props
	}

	appDetails, err := deserializeClientApp(appDetailsMap)
	if err != nil {
		return nil, err
	}

	return appDetails, nil
}

func (c ClientAppStorage) GetAll(opts ...string) ([]*domain.ClientApp, error) {

	return nil, nil
}

func (c ClientAppStorage) Get(clientId string) (*domain.ClientApp, error) {
	records, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (a:ClientApplication)
				WHERE a.id = $clientId
				RETURN a
   		`

		result, err := tx.Run(cypher, map[string]interface{}{"clientId": clientId})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		return nil, err
	}

	var clientMap map[string]interface{}

	for _, record := range records.([]*neo4j.Record) {
		clientMap = record.Values[0].(neo4j.Node).Props
	}

	clientApp, err := deserializeClientApp(clientMap)

	if err != nil {
		return &domain.ClientApp{}, err
	}

	return clientApp, nil
}

func (c ClientAppStorage) Delete(clientId string) error {
	return nil
}

func (c ClientAppStorage) AuthorizeClientCredentials(requestToken, clientId string) (*domain.ClientApp, error) {
	records, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		authorizeClientQuery := `
				MATCH (a:ClientApplication)
				WHERE a.id = $clientId AND a.requestToken = $requestToken
				RETURN a
   		`

		result, err := tx.Run(authorizeClientQuery, map[string]interface{}{"clientId": clientId, "requestToken": requestToken})
		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		return nil, err
	}

	var clientMap map[string]interface{}

	for _, record := range records.([]*neo4j.Record) {
		clientMap = record.Values[0].(neo4j.Node).Props
	}

	clientApp, err := deserializeClientApp(clientMap)

	if err != nil {
		return &domain.ClientApp{}, err
	}

	return clientApp, nil
}
