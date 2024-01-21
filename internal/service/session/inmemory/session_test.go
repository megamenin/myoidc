package inmemory

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManager_Create(t *testing.T) {
	mgr := NewManager()

	sess, err := mgr.Create(context.TODO(), "123", map[string]interface{}{
		"field1": 1,
		"field2": "2",
	})

	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 1, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess, mgr.store[sess.Id], "sessions don't much")
	assert.NotEmpty(t, sess.Id, "invalid session id")
	assert.Equal(t, sess.UserId, "123", "invalid user id")
	assert.Equal(t, sess.Data, map[string]interface{}{
		"field1": 1,
		"field2": "2",
	}, "invalid session data")
}

func TestManager_CreateTemp(t *testing.T) {
	mgr := NewManager()

	sess, err := mgr.CreateTemp(context.TODO(), map[string]interface{}{
		"field1": 1,
		"field2": "2",
	})

	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 1, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess, mgr.store[sess.Id], "sessions don't much")
	assert.NotEmpty(t, sess.Id, "invalid session id")
	assert.Empty(t, sess.UserId, "invalid user id for temp session")
	assert.Equal(t, sess.Data, map[string]interface{}{
		"field1": 1,
		"field2": "2",
	}, "invalid session data")
}

func TestManager_Get(t *testing.T) {
	mgr := NewManager()

	sess1, err := mgr.CreateTemp(context.TODO(), map[string]interface{}{
		"field1": 11,
		"field2": "22",
	})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 1, len(mgr.store), "sessions count is invalid")
	sess2, err := mgr.Create(context.TODO(), "123", map[string]interface{}{
		"field1": 12,
		"field2": "22",
	})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 2, len(mgr.store), "sessions count is invalid")

	sess1res, err := mgr.Get(context.TODO(), sess1.Id)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 2, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess1, sess1res, "sessions don't much")

	sess2res, err := mgr.Get(context.TODO(), sess2.Id)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 2, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess2, sess2res, "sessions don't much")

	_, err = mgr.Get(context.TODO(), "not_exists")
	assert.Error(t, err, "error expected")
}

func TestManager_Destroy(t *testing.T) {
	mgr := NewManager()

	sess1, err := mgr.CreateTemp(context.TODO(), map[string]interface{}{
		"field1": 11,
		"field2": "22",
	})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 1, len(mgr.store), "sessions count is invalid")
	sess2, err := mgr.Create(context.TODO(), "123", map[string]interface{}{
		"field1": 12,
		"field2": "22",
	})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 2, len(mgr.store), "sessions count is invalid")

	err = mgr.Destroy(context.TODO(), "not_exists")
	assert.NoError(t, err, "unexpected error")

	sess1res, err := mgr.Get(context.TODO(), sess1.Id)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 2, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess1, sess1res, "sessions don't much")

	sess2res, err := mgr.Get(context.TODO(), sess2.Id)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 2, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess2, sess2res, "sessions don't much")

	err = mgr.Destroy(context.TODO(), sess1.Id)
	assert.NoError(t, err, "unexpected error")

	sess1res, err = mgr.Get(context.TODO(), sess1.Id)
	assert.Error(t, err, "error expected")
	assert.Equal(t, 1, len(mgr.store), "sessions count is invalid")

	sess2res, err = mgr.Get(context.TODO(), sess2.Id)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, 1, len(mgr.store), "sessions count is invalid")
	assert.Equal(t, sess2, sess2res, "sessions don't much")

	err = mgr.Destroy(context.TODO(), sess2.Id)
	assert.NoError(t, err, "unexpected error")

	sess2res, err = mgr.Get(context.TODO(), sess2.Id)
	assert.Error(t, err, "error expected")
	assert.Equal(t, 0, len(mgr.store), "sessions count is invalid")
}
