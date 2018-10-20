package db

import (
	"errors"

	"github.com/iampigeon/pigeon"
)

// ChannelStore ...
type ChannelStore struct {
	Dst *Datastore
}

// GetChannels ...
func (cs *ChannelStore) GetChannels() ([]*pigeon.Channel, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	return mock.Channels, nil
}

// GetChannelById ...
func (cs *ChannelStore) GetChannelById(id string) (*pigeon.Channel, error) {
	mock, err := getMock()
	if err != nil {
		return nil, err
	}

	for _, value := range mock.Channels {
		if value.ID == id {
			return value, nil
		}
	}

	return nil, errors.New("channel not found")
}
