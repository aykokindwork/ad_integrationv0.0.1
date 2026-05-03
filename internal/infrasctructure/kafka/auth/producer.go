package kafkaauth

import (
	"ad_integration/internal/auth/service"
	"context"
	"fmt"
)

const topicAuthLog = "auth-log"

type AuthProducer struct {
	base service.EventProducer
}

func NewAuthProducer(base service.EventProducer) *AuthProducer {
	return &AuthProducer{
		base: base,
	}
}

func (p *AuthProducer) PublishLDAPSuccess(ctx context.Context, email string) error {
	return p.publish(ctx, email, AuthEvent{
		Username: email,
		Event:    EventLDAPSuccess,
		Status:   "AWAITING_OTP",
	})
}

func (p *AuthProducer) PublishOTPSuccess(ctx context.Context, email string) error {
	return p.publish(ctx, email, AuthEvent{
		Username: email,
		Event:    EventOTPSuccess,
		Status:   "AUTHENTICATED",
	})
}

func (p *AuthProducer) PublishAuthFailed(ctx context.Context, email string, reason string) error {
	return p.publish(ctx, email, AuthEvent{
		Username: email,
		Event:    EventAuthFailed,
		Status:   reason,
	})
}

func (p *AuthProducer) publish(ctx context.Context, key string, event AuthEvent) error {
	payload, err := event.Encode()
	if err != nil {
		return fmt.Errorf("kafka: encode failed: %w", err)
	}
	return p.base.SendMessage(ctx, topicAuthLog, []byte(key), payload)
}
