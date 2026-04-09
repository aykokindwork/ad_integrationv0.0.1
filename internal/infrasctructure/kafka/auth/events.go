package kafkaauth

import "encoding/json"

type AuthEventType string

const (
	EventLDAPSuccess AuthEventType = "LDAP_AUTH_SUCCESS"
	EventOTPSuccess  AuthEventType = "OTP_AUTH_SUCCESS"
	EventAuthFailed  AuthEventType = "AUTH_FAILED"
)

type AuthEvent struct {
	Username string        `json:"username"`
	Event    AuthEventType `json:"event"`
	Status   string        `json:"status"`
}

func (e AuthEvent) Encode() ([]byte, error) {
	return json.Marshal(e)
}
