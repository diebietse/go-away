package firealert

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

type storage interface {
	add(context.Context, string, Alert) error
	delete(context.Context, string) error
}

type storeClient struct {
	collection *firestore.CollectionRef
}

func newStore(ctx context.Context, app *firebase.App) (*storeClient, error) {
	store, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %v", err)
	}

	c := store.Collection(FirebaseTopic)
	if c == nil {
		return nil, fmt.Errorf("could not resolve collection for %v", FirebaseTopic)
	}

	return &storeClient{
		collection: c,
	}, nil
}

func (s *storeClient) add(ctx context.Context, ID string, data Alert) error {
	doc := s.collection.Doc(ID)
	_, err := doc.Create(ctx, data)
	if err != nil {
		return fmt.Errorf("could not create document: %v", err)
	}
	return nil
}

func (s *storeClient) delete(ctx context.Context, ID string) error {
	doc := s.collection.Doc(ID)
	_, err := doc.Delete(ctx)
	if err != nil {
		return fmt.Errorf("could not delete document: %v", err)
	}
	return nil
}
