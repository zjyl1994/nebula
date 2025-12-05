package util

import "golang.org/x/sync/singleflight"

type Singleflight[T any] struct {
	group singleflight.Group
}

func (s *Singleflight[T]) Do(key string, fn func() (T, error)) (value T, err error) {
	ret, err, _ := s.group.Do(key, func() (interface{}, error) {
		data, e := fn()
		return data, e
	})
	if err != nil {
		return
	}
	return ret.(T), nil
}
