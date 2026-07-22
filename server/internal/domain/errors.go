// Package domain 承载业务概念：核心类型、语义化错误与身份信息。
package domain

import (
	"errors"
	"fmt"
)

// Kind 错误分类，对应 HTTP 状态码。
type Kind int

const (
	KindValidation Kind = iota + 1
	KindUnauthed
	KindForbidden
	KindNotFound
	KindConflict
	KindInternal
)

// Error 语义化业务错误。
type Error struct {
	Kind    Kind
	Message string
	Wrapped error
}

func (e *Error) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Wrapped)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Wrapped }

// Wrap 将底层错误包装为领域错误。
func (e *Error) Wrap(err error) *Error {
	return &Error{Kind: e.Kind, Message: e.Message, Wrapped: err}
}

var (
	ErrNamespaceNotFound  = &Error{Kind: KindNotFound, Message: "namespace not found"}
	ErrConfigNotFound     = &Error{Kind: KindNotFound, Message: "config not found"}
	ErrHistoryNotFound    = &Error{Kind: KindNotFound, Message: "config history not found"}
	ErrAccessKeyNotFound  = &Error{Kind: KindNotFound, Message: "access key not found"}
	ErrPermissionNotFound = &Error{Kind: KindNotFound, Message: "permission not found"}
	ErrConflict           = &Error{Kind: KindConflict, Message: "optimistic lock conflict"}
	ErrDuplicate          = &Error{Kind: KindConflict, Message: "duplicate entry"}
	ErrUnauthenticated    = &Error{Kind: KindUnauthed, Message: "unauthenticated"}
	ErrForbidden          = &Error{Kind: KindForbidden, Message: "forbidden"}
	ErrInternal           = &Error{Kind: KindInternal, Message: "internal error"}
	ErrValidation         = &Error{Kind: KindValidation, Message: "validation error"}
	ErrNamespaceNotEmpty  = &Error{Kind: KindConflict, Message: "namespace is not empty, delete its configs and permissions first"}
)

// MapError 将仓储错误转为领域错误。
func MapError(err error) error {
	if err == nil {
		return nil
	}
	var de *Error
	if errors.As(err, &de) {
		return err
	}
	return ErrInternal.Wrap(err)
}

// Identity 经过认证的用户身份。
type Identity struct {
	Owner       string
	Username    string
	DisplayName string
	Email       string
	Phone       string
	Avatar      string
	Roles       []string
	Groups      []string
	Properties  map[string]string
}
