package main

import (
	"errors"
	"sync"
)

type InMemoryUserStorage struct {
	lock	sync.RWMutex
	storage	map[string]User
}

func NewInMemoryUserStorage() *InMemoryUserStorage {
	return &InMemoryUserStorage{
		lock: sync.RWMutex{},
		storage: make(map[string]User),
	}
}

func (repository *InMemoryUserStorage) Add(key string, usr User) error {
	repository.lock.Lock()
	defer repository.lock.Unlock()
	if _, ok := repository.storage[key]; ok {
		return errors.New("The user already exists")
	}

	repository.storage[key] = usr
	return nil
}

// Add should return error if user with given key (login) is already present


func (repository *InMemoryUserStorage) Update(key string, usr User) error {
	repository.lock.Lock()
	defer repository.lock.Unlock()

	if _, ok := repository.storage[key]; !ok {
		return errors.New("The user doesn't exist")
	}
	repository.storage[key] = usr

	return nil
}

// Update should return error if there is no such user to update


func (repository *InMemoryUserStorage) Get(key string) (User, error) {
	repository.lock.Lock()
	defer repository.lock.Unlock()
	var returnValue User

	if _, ok := repository.storage[key]; !ok {
		return returnValue, errors.New("The user doesn't exist")
	}
	returnValue = repository.storage[key]
	return returnValue, nil

}


func (repository *InMemoryUserStorage) Delete(key string) (User, error) {
	repository.lock.Lock()
	defer repository.lock.Unlock()
	var returnValue User

	if _, ok := repository.storage[key]; !ok {
		return returnValue, errors.New("The user doesn't exist")
	}
	returnValue = repository.storage[key]
	delete(repository.storage, key)
	return returnValue, nil
}

// Delete should return error if there is no such user to delete
// Delete should return deleted user
