package firealert

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
)

type messenger interface {
	send(context.Context, *messaging.Message) error
}

type msgClient struct {
	msg *messaging.Client
}

func newMsg(ctx context.Context, app *firebase.App) (*msgClient, error) {
	m, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing messaging: %v", err)
	}

	return &msgClient{
		msg: m,
	}, nil
}

func (m *msgClient) send(ctx context.Context, message *messaging.Message) error {
	_, err := m.msg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("could not send message: %v", err)
	}
	return nil
}
