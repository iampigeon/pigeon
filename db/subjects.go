package db

import (
	"errors"

	"github.com/iampigeon/pigeon"
)

// SubjectStore ...
type SubjectStore struct {
	*Datastore
}

// GetSubjects ...
func (ss *SubjectStore) GetSubjects() ([]*pigeon.Subject, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	return mock.Subjects, nil
}

// AddSubject ...
func (ss *SubjectStore) AddSubject(m *pigeon.Subject) error {
	//save subject
	return nil
}

// GetSubjectsByUserID ...
func (ss *SubjectStore) GetSubjectsByUserID(userID string) ([]*pigeon.Subject, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	subjects := make([]*pigeon.Subject, 0)

	for _, subject := range mock.Subjects {
		if subject.UserID == userID {
			subjects = append(subjects, subject)
		}
	}

	return subjects, nil
}

// GetUserSubjectByName ...
func (ss *SubjectStore) GetUserSubjectByName(userID string, name string) (*pigeon.Subject, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	for _, subject := range mock.Subjects {
		if subject.UserID == userID && subject.Name == name {
			return subject, nil
		}
	}

	return nil, errors.New("subject not found")
}
