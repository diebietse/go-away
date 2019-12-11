package firealert

import (
	"context"
	"testing"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/stretchr/testify/assert"
)

const (
	testTitle = "Title"
	testID    = "id"
	testBody  = "This is a body"
)

type mockInterface struct {
	IDAdd    string
	IDDelete string
	Data     *Alert
	Msg      *messaging.Message
}

func mock() *mockInterface {
	return &mockInterface{}
}

func (m *mockInterface) add(ctx context.Context, ID string, data Alert) error {
	m.IDAdd = ID
	m.Data = &data
	return nil
}

func (m *mockInterface) delete(ctx context.Context, ID string) error {
	m.IDDelete = ID
	return nil
}

func (m *mockInterface) send(ctx context.Context, message *messaging.Message) error {
	m.Msg = message
	return nil
}

func mockFire(m *mockInterface) *Firealert {
	return &Firealert{
		ctx:   context.Background(),
		msg:   m,
		store: m,
		lvl:   LevelWarning,
	}
}

func compareMsg(t *testing.T, m *mockInterface, msg *messaging.Message) {
	assert.NotNil(t, msg, "Message is nil")
	assert.NotNil(t, msg.Notification, "Notification is nil")
	assert.Equal(t, m.Data.Title, msg.Notification.Title, "Message titles do no match")
	assert.Equal(t, m.Data.Body, msg.Notification.Body, "Message bodies do no match")
}

func TestAddWarning(t *testing.T) {
	testTime := time.Now()
	ctx := context.Background()
	m := mock()
	f := mockFire(m)
	a := Alert{
		ID:    testID,
		Title: testTitle,
		Body:  testBody,
		Time:  testTime,
		Level: LevelWarning,
		State: StateTriggered,
	}

	err := f.SendAlert(ctx, a)
	assert.NoError(t, err, "SendAlert created an error")
	assert.Equal(t, a.ID, m.IDAdd, "SendAlert ID does not match")
	assert.Equal(t, &a, m.Data, "Send ID data does not match")
	assert.Equal(t, "", m.IDDelete, "ID should be nil")
	compareMsg(t, m, m.Msg)
}

func TestRemoveWarning(t *testing.T) {
	testTime := time.Now()
	ctx := context.Background()
	m := mock()
	f := mockFire(m)
	a := Alert{
		ID:    testID,
		Title: testTitle,
		Body:  testBody,
		Time:  testTime,
		Level: LevelWarning,
		State: StateResolved,
	}

	err := f.SendAlert(ctx, a)
	assert.NoError(t, err, "SendAlert created an error")
	assert.Equal(t, "", m.IDAdd, "ID should be nil")
	assert.Nil(t, m.Data, "No data should be written")
	assert.Equal(t, a.ID, m.IDDelete, "SendAlert ID does not match")
	assert.Nil(t, m.Msg, "No data should be written")
}

func TestAddInfo(t *testing.T) {
	testTime := time.Now()
	ctx := context.Background()
	m := mock()
	f := mockFire(m)
	a := Alert{
		ID:    testID,
		Title: testTitle,
		Body:  testBody,
		Time:  testTime,
		Level: LevelInfo,
		State: StateTriggered,
	}

	err := f.SendAlert(ctx, a)
	assert.NoError(t, err, "SendAlert created an error")
	assert.Equal(t, a.ID, m.IDAdd, "SendAlert ID does not match")
	assert.Equal(t, &a, m.Data, "Send ID data does not match")
	assert.Equal(t, "", m.IDDelete, "ID should be nil")
	assert.Nil(t, m.Msg, "No message should be sent")
}

func TestRemoveInfo(t *testing.T) {
	testTime := time.Now()
	ctx := context.Background()
	m := mock()
	f := mockFire(m)
	a := Alert{
		ID:    testID,
		Title: testTitle,
		Body:  testBody,
		Time:  testTime,
		Level: LevelInfo,
		State: StateResolved,
	}

	err := f.SendAlert(ctx, a)
	assert.NoError(t, err, "SendAlert created an error")
	assert.Equal(t, "", m.IDAdd, "ID should be nil")
	assert.Nil(t, m.Data, "No data should be written")
	assert.Equal(t, a.ID, m.IDDelete, "SendAlert ID does not match")
	assert.Nil(t, m.Msg, "No data should be written")
}

func TestInvalidState(t *testing.T) {
	testTime := time.Now()
	ctx := context.Background()
	m := mock()
	f := mockFire(m)
	a := Alert{
		ID:    testID,
		Title: testTitle,
		Body:  testBody,
		Time:  testTime,
		Level: LevelInfo,
		State: 0,
	}
	err := f.SendAlert(ctx, a)
	assert.Error(t, err, "Invalid state was not catched")
}
