package connections

import (
	"context"
	"sync"
	"testing"

	"cloud.google.com/go/pubsub"
)

type fakeTopic struct {
	mu        sync.Mutex
	exists    int
	published int
	stopped   int
}

func (f *fakeTopic) Exists(context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.exists++
	return true, nil
}

func (f *fakeTopic) Publish(context.Context, *pubsub.Message) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.published++
	return "id", nil
}

func (f *fakeTopic) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped++
}

func TestTopicCache_ExistsOnce(t *testing.T) {
	ft := &fakeTopic{}
	opened := 0
	g := &gPubSubWrapper{
		topics:    map[string]topicHandle{},
		openTopic: func(string) topicHandle { opened++; return ft },
	}

	for i := 0; i < 50; i++ {
		tp, err := g.topic(context.Background(), "order-import-events")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tp.Publish(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}
	if ft.exists != 1 || opened != 1 || ft.published != 50 {
		t.Fatalf("exists=%d opened=%d published=%d, want 1/1/50", ft.exists, opened, ft.published)
	}

	g.stopTopics()
	if ft.stopped != 1 {
		t.Fatalf("stopped=%d, want 1", ft.stopped)
	}
}
