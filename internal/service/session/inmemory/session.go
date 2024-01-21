package inmemory

import (
	"context"
	"github.com/google/uuid"
	"myoidc/internal/service/session"
	"myoidc/pkg/errors"
	"sync"
)

// Manager simple in-memory implementation of session.Manager interface.
type Manager struct {
	store map[string]*session.Session
	mx    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		store: make(map[string]*session.Session),
		mx:    sync.RWMutex{},
	}
}

func (mgr *Manager) Create(ctx context.Context, userId string, data map[string]interface{}) (*session.Session, error) {
	return mgr.create(userId, data)
}

func (mgr *Manager) CreateTemp(ctx context.Context, data map[string]interface{}) (*session.Session, error) {
	return mgr.create("", data)
}

func (mgr *Manager) create(userId string, data map[string]interface{}) (*session.Session, error) {
	sessUuid, err := uuid.NewUUID()
	if err != nil {
		err = errors.Wrap(err, "failed to generate session id")
		return nil, err
	}

	sessId := sessUuid.String()
	sess := session.NewSession(sessId, userId, data)

	mgr.mx.Lock()
	mgr.store[sessId] = sess
	mgr.mx.Unlock()

	return sess, nil
}

func (mgr *Manager) Get(ctx context.Context, sessId string) (*session.Session, error) {
	mgr.mx.RLock()
	sess := mgr.store[sessId]
	mgr.mx.RUnlock()

	if sess == nil {
		err := errors.Errorf("session not found by id %s", sessId)
		return nil, err
	}

	return sess, nil
}

func (mgr *Manager) Destroy(ctx context.Context, sessId string) error {
	mgr.mx.Lock()
	delete(mgr.store, sessId)
	mgr.mx.Unlock()

	return nil
}
