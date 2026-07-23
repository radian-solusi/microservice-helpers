package connections

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
)

type gPubSubWrapper struct {
	client           *pubsub.Client
	ctx              context.Context
	projectID        string
	cancelFuncs      map[string]context.CancelFunc
	mu               sync.Mutex
	messagesReceived int64
	messagesAcked    int64
	messagesNacked   int64
	lastMessageTime  time.Time
}

func NewGPubSub(ctx context.Context, cfg helperconfig.GPubSubConfig) (GPubSub, error) {
	wrapper := &gPubSubWrapper{
		ctx:         ctx,
		cancelFuncs: make(map[string]context.CancelFunc),
	}

	if cfg.ProjectID == "" {
		return wrapper, nil
	}

	var client *pubsub.Client
	var err error

	if cfg.EmulatorHost != "" {
		client, err = pubsub.NewClient(ctx, cfg.ProjectID,
			option.WithEndpoint(cfg.EmulatorHost),
			option.WithoutAuthentication(),
		)
	} else if cfg.CredentialsFile != "" {
		client, err = pubsub.NewClient(ctx, cfg.ProjectID, option.WithCredentialsFile(cfg.CredentialsFile))
	} else {
		client, err = pubsub.NewClient(ctx, cfg.ProjectID)
	}
	if err != nil {
		return wrapper, fmt.Errorf("create Google Pub/Sub client: %w", err)
	}

	wrapper.client = client
	wrapper.projectID = cfg.ProjectID
	return wrapper, nil
}

func (g *gPubSubWrapper) Client() *pubsub.Client { return g.client }

func (g *gPubSubWrapper) IsConnected() bool {
	if g.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(g.ctx, 5*time.Second)
	defer cancel()
	topic := g.client.Topic("connection_test")
	_, err := topic.Exists(ctx)
	return err == nil
}

func (g *gPubSubWrapper) Publish(ctx context.Context, topicID string, data []byte, attributes map[string]string) (string, error) {
	if g.client == nil {
		return "", errors.New("pubsub client not connected")
	}
	topic := g.client.Topic(topicID)
	defer topic.Stop()

	exists, err := topic.Exists(ctx)
	if err != nil {
		return "", fmt.Errorf("check topic existence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("topic %s does not exist", topicID)
	}

	result := topic.Publish(ctx, &pubsub.Message{Data: data, Attributes: attributes})
	messageID, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("publish message: %w", err)
	}
	return messageID, nil
}

func (g *gPubSubWrapper) Subscribe(ctx context.Context, subscriptionID string, handler func(msg *pubsub.Message)) error {
	if g.client == nil {
		return errors.New("pubsub client not connected")
	}
	sub := g.client.Subscription(subscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute
	sub.ReceiveSettings.NumGoroutines = 10

	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check subscription existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("subscription %s does not exist", subscriptionID)
	}

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		atomic.AddInt64(&g.messagesReceived, 1)
		g.lastMessageTime = time.Now()
		handler(msg)
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("subscription error: %w", err)
	}
	return nil
}

func (g *gPubSubWrapper) SubscribeAsync(ctx context.Context, subscriptionID string, handler func(msg *pubsub.Message)) error {
	if g.client == nil {
		return errors.New("pubsub client not connected")
	}
	sub := g.client.Subscription(subscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute
	sub.ReceiveSettings.NumGoroutines = 10

	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check subscription existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("subscription %s does not exist", subscriptionID)
	}

	subCtx, cancel := context.WithCancel(ctx)
	g.mu.Lock()
	g.cancelFuncs[subscriptionID] = cancel
	g.mu.Unlock()

	go func() {
		_ = sub.Receive(subCtx, func(ctx context.Context, msg *pubsub.Message) {
			atomic.AddInt64(&g.messagesReceived, 1)
			g.lastMessageTime = time.Now()
			handler(msg)
		})
	}()
	return nil
}

func (g *gPubSubWrapper) StopSubscription(subscriptionID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	cancel, ok := g.cancelFuncs[subscriptionID]
	if !ok {
		return fmt.Errorf("subscription %s not found or not started with SubscribeAsync", subscriptionID)
	}
	cancel()
	delete(g.cancelFuncs, subscriptionID)
	return nil
}

func (g *gPubSubWrapper) CreateTopic(ctx context.Context, topicID string) error {
	if g.client == nil {
		return errors.New("pubsub client not connected")
	}
	topic := g.client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check topic existence: %w", err)
	}
	if exists {
		return nil
	}
	_, err = g.client.CreateTopic(ctx, topicID)
	if err != nil {
		return fmt.Errorf("create topic: %w", err)
	}
	return nil
}

func (g *gPubSubWrapper) CreateSubscription(ctx context.Context, topicID string, subscriptionID string) error {
	if g.client == nil {
		return errors.New("pubsub client not connected")
	}
	sub := g.client.Subscription(subscriptionID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check subscription existence: %w", err)
	}
	if exists {
		return nil
	}
	_, err = g.client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic: g.client.Topic(topicID),
	})
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (g *gPubSubWrapper) TopicExists(ctx context.Context, topicID string) (bool, error) {
	if g.client == nil {
		return false, errors.New("pubsub client not connected")
	}
	return g.client.Topic(topicID).Exists(ctx)
}

func (g *gPubSubWrapper) SubscriptionExists(ctx context.Context, subscriptionID string) (bool, error) {
	if g.client == nil {
		return false, errors.New("pubsub client not connected")
	}
	return g.client.Subscription(subscriptionID).Exists(ctx)
}

func (g *gPubSubWrapper) DeleteTopic(ctx context.Context, topicID string) error {
	if g.client == nil {
		return errors.New("pubsub client not connected")
	}
	return g.client.Topic(topicID).Delete(ctx)
}

func (g *gPubSubWrapper) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	if g.client == nil {
		return errors.New("pubsub client not connected")
	}
	return g.client.Subscription(subscriptionID).Delete(ctx)
}

func (g *gPubSubWrapper) GetStats() SubscriptionStats {
	return SubscriptionStats{
		MessagesReceived: atomic.LoadInt64(&g.messagesReceived),
		MessagesAcked:    atomic.LoadInt64(&g.messagesAcked),
		MessagesNacked:   atomic.LoadInt64(&g.messagesNacked),
		LastMessageTime:  g.lastMessageTime,
	}
}

func (g *gPubSubWrapper) Close() error {
	if g.client == nil {
		return nil
	}
	g.mu.Lock()
	for id, cancel := range g.cancelFuncs {
		cancel()
		delete(g.cancelFuncs, id)
	}
	g.mu.Unlock()
	return g.client.Close()
}
