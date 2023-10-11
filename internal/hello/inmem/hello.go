package inmem

import (
	"fmt"
	e "go-misc/internal/errors"
	"go-misc/internal/hello"
	"sync"
)

type HelloRepository struct {
	mtx      sync.RWMutex
	messages map[string]*hello.Message
}

// https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ hello.HelloRepository = (*HelloRepository)(nil)

func NewRepository() *HelloRepository {
	ms := []*hello.Message{
		{Id: "hello", English: "Hello", French: "Bonjour", Malagasy: "Salama"},
		{Id: "goodbye", English: "Good bye", French: "Au revoir", Malagasy: "Veloma"},
	}
	r := &HelloRepository{messages: make(map[string]*hello.Message)}
	for _, m := range ms {
		r.messages[m.Id] = m
	}
	return r
}

func (r *HelloRepository) Get(id string) (hello.Message, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	m, ok := r.messages[id]
	if !ok {
		return hello.Message{}, fmt.Errorf("get message %q: %w", id, e.New(e.CodeNotFound, "message not found"))
	}

	return *m, nil
}
