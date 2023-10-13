package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/yenonn/pcbook/pb"
)

var ErrAlreadyExists = errors.New("record is exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, sendStream func(laptop *pb.Laptop) error) error
}

type InMemoryLaptopStore struct {
	data  map[string]*pb.Laptop
	mutex sync.RWMutex
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{data: make(map[string]*pb.Laptop)}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		// raise problem
		return ErrAlreadyExists
	}
	// deep copy
	other, err := deepCopy(laptop)
	if err != nil {
		return fmt.Errorf("cannot copy laptop data %w", err)
	}
	store.data[other.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}
	// found
	other, err := deepCopy(laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data, %w", err)
	}

	return other, nil
}

func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, sendStream func(laptop *pb.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	for _, laptop := range store.data {
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			log.Print("context is cancelled.")
			return errors.New("context is cancelled")
		}
		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			err = sendStream(other)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
