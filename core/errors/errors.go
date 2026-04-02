package errors

import (
	"fmt"
)

type AppError struct {
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithErr(err error) error {
	return &AppError{
		Message: e.Message,
		Err:     err,
	}
}

var (
	ErrInvalidCredentials = &AppError{Message: "invalid credentials"}
	ErrUserIsBlocked      = &AppError{Message: "user is blocked"}

	//ldapErrs
	ErrLdapUnexpected = &AppError{Message: "failed to LDAP Auth"}
	ErrLdapTLS        = &AppError{Message: "failed to connect TLS"}

	ErrLdapSearch       = &AppError{Message: "failed to search attributes"}
	ErrLdapNoDN         = &AppError{Message: "not saved DN"}
	ErrLdapNoAttributes = &AppError{Message: "not saved attributes"}
	ErrAdUserNotFound   = &AppError{Message: "user is not found"}
	ErrLdapNoEmail      = &AppError{Message: "not saved email"}
	ErrLdapNoName       = &AppError{Message: "not saved name"}

	ErrLdapBind = &AppError{Message: "failed to bind user"}
)
