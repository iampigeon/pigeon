package db

import (
	"errors"
	"time"

	"github.com/iampigeon/pigeon"
)

// CriteriaStore ...
type CriteriaStore struct {
	Dst *Datastore
}

// GetCriterias ...
func (ts *CriteriaStore) GetCriterias() ([]*pigeon.Criteria, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	return mock.Criterias, nil
}

// GetCriteriaById ...
func (ts *CriteriaStore) GetCriteriaById(ID string) (*pigeon.Criteria, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	for _, value := range mock.Criterias {
		if value.ID == ID {
			return value, nil
		}
	}

	return nil, errors.New("criteria not found")
}

// GetCriteriaDelay ...
func (ts *CriteriaStore) GetCriteriaDelay(criteriaID string, criteriaCustom int64) (time.Duration, error) {
	criteria, err := ts.GetCriteriaById(criteriaID)
	if err != nil {
		return 0, err
	}

	var delay time.Duration

	if criteria.Value == -1 {
		delay = time.Duration(criteriaCustom) * time.Second
	} else {
		delay = time.Duration(criteria.Value) * time.Second
	}

	return delay, nil
}
