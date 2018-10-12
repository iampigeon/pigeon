package db

import (
	"errors"

	"github.com/iampigeon/pigeon"
)

// UserStore ...
type UserStore struct {
	*Datastore
}

// GetUsers ...
func (us *UserStore) GetUsers() ([]*pigeon.User, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	return mock.Users, nil
}

// GetUserByAPIKey ...
func (us *UserStore) GetUserByAPIKey(APIKey string) (*pigeon.User, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	var user *pigeon.User

	for _, value := range mock.Users {
		if value.APIKey == APIKey {
			user = value
			continue
		}
	}

	if user == nil {
		return nil, errors.New("user not found or invalid api key")
	}

	return user, nil
}
