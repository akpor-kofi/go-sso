package neo4j

import (
	"go-sso/internal/core/domain"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type UserStorage struct {
}

func NewUserStorage() *UserStorage {
	return &UserStorage{}
}

func (u UserStorage) GetAll() ([]*domain.User, error) {
	result, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (u:User)
				RETURN u
   		`

		result, err := tx.Run(cypher, map[string]interface{}{})

		if err != nil {
			return nil, err
		}

		users := make([]map[string]interface{}, 0) // len doesn't really matter

		for result.Next() {
			user := result.Record().Values[0].(neo4j.Node).Props

			users = append(users, user)

		}

		return deserializeUsers(users), result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]*domain.User), nil

}

func (u UserStorage) Get(id string) (*domain.User, error) {
	records, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (u:User)
				WHERE u.id = $userId
				RETURN u
   		`

		result, err := tx.Run(cypher, map[string]interface{}{"userId": id})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	var userMap map[string]interface{}

	for _, record := range records.([]*neo4j.Record) {
		userMap = record.Values[0].(neo4j.Node).Props
	}

	user, err := deserializeUser(userMap)

	if err != nil {
		return &domain.User{}, err
	}

	return user, nil
}

func (u UserStorage) GetByEmail(email string) (*domain.User, error) {
	records, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (u:User)
				WHERE u.email = $email
				RETURN u
   		`

		result, err := tx.Run(cypher, map[string]interface{}{"email": email})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	var userMap map[string]interface{}

	for _, record := range records.([]*neo4j.Record) {
		userMap = record.Values[0].(neo4j.Node).Props
	}

	user, err := deserializeUser(userMap)

	if err != nil {
		return &domain.User{}, err
	}

	return user, nil
}

func (u UserStorage) New(user *domain.User) (*domain.User, error) {
	properties := serializeDataToMap(user)

	_, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
			CREATE (u:User {id: $Id, name: $Name, email: $Email, image: $Image, dob: $Dob, password: $Password, employeeId: $EmployeeId, phoneNumber: $PhoneNumber, createdAt: $CreatedAt, updatedAt: $UpdatedAt, version: $Version, resetToken: $ResetToken, resetExpiresAt: $ResetExpiresAt})
`

		// uniqueConstraints := `CREATE CONSTRAINT user_email_unique IF NOT EXISTS FOR (user:User) REQUIRE user.email IS UNIQUE`

		result, err := tx.Run(cypher, properties)
		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	return user, nil

}

func (u UserStorage) Update(id string, user *domain.User) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserStorage) Delete(id string) error {
	_, err := Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
				MATCH (u:User)
				WHERE u.id = $userId
				DELETE u
   		`

		result, err := tx.Run(cypher, map[string]interface{}{"userId": id})
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

func (u UserStorage) UpdateResetToken(email, resetToken string) error {
	_, err := Session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		cypher := `
			MATCH(u:User {email: $email})
			SET u.resetToken = $resetToken
			SET u.resetExpiresAt = $exp
		`

		result, err := tx.Run(cypher, map[string]interface{}{
			"email":      email,
			"resetToken": resetToken,
			"exp":        time.Now().Add(10 * time.Minute).UnixMilli(),
		})

		if err != nil {
			return nil, err
		}

		return result.Collect()
	})

	if err != nil {
		panic(err)
	}

	return nil
}
