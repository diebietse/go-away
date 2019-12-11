package firealert

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

const FirebaseTopic = "alert"

type AlertLevel int

const (
	LevelUnknown = AlertLevel(iota)
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
)

var AlertStrings = map[AlertLevel]string{
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarning: "WARNING",
	LevelError:   "ERROR",
}

type AlertState int

const (
	StateUnknown = AlertState(iota)
	StateResolved
	StateTriggered
)

type Firealert struct {
	ctx   context.Context
	msg   messenger
	store storage
	lvl   AlertLevel
}

type Alert struct {
	ID    string     `json:"-"`
	Title string     `json:"title"`
	Body  string     `json:"body"`
	Time  time.Time  `json:"timestamp"`
	Level AlertLevel `json:"level"`
	State AlertState `json:"-"`
}

func New(configPath string, lvl AlertLevel) (*Firealert, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	store, err := newStore(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %v", err)
	}

	msg, err := newMsg(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("error initializing messaging: %v", err)
	}

	return &Firealert{
		ctx:   ctx,
		msg:   msg,
		store: store,
		lvl:   lvl,
	}, nil
}

func (f *Firealert) SendAlert(ctx context.Context, alert Alert) error {
	if err := f.handleStore(ctx, alert); err != nil {
		return err
	}

	if err := f.handleMessage(ctx, alert); err != nil {
		return err
	}

	return nil
}

func (f *Firealert) handleStore(ctx context.Context, alert Alert) error {
	var err error
	switch alert.State {
	case StateResolved:
		err = f.store.delete(ctx, alert.ID)
	case StateTriggered:
		err = f.store.add(ctx, alert.ID, alert)
	default:
		err = fmt.Errorf("alert has unknown state: %v", alert.State)
	}
	return err
}

func (f *Firealert) handleMessage(ctx context.Context, alert Alert) error {
	if alert.Level < f.lvl {
		return nil
	}
	if alert.State != StateTriggered {
		return nil
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: alert.Title,
			Body:  alert.Body,
		},
		Topic: FirebaseTopic,
	}

	return f.msg.send(ctx, message)
}
