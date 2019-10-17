package serving

import (
	"fmt"
)

//revive:disable:exported

type ValueIndex map[string]uint32

type KVstore struct {
	store map[FeatureName]ValueIndex
	uniqs map[FeatureName]uint32
}

func NewKVStore() *KVstore {
	return &KVstore{
		store: make(map[FeatureName]ValueIndex),
		uniqs: make(map[FeatureName]uint32),
	}
}

func (s *KVstore) Set(key FeatureName, val string) (uint32, error) {
	inner, ok := s.store[key]
	if !ok {
		newMap := make(map[string]uint32)
		newMap[val] = 0
		s.store[key] = newMap
		s.uniqs[key] = 1
		return 0, nil
	}

	no, ok := inner[val]
	if !ok {
		max := s.uniqs[key]
		inner[val] = max
		s.uniqs[key] = max + 1
		return max, nil
	}

	return no, nil
}

func (s *KVstore) Get(key FeatureName, val string) (uint32, error) {
	inner, ok := s.store[key]
	if !ok {
		return 0, fmt.Errorf("Missing feature %v", key)
	}

	no, ok := inner[val]
	if !ok {
		return 0, fmt.Errorf("Missing value %v", val)
	}

	return no, nil
}
