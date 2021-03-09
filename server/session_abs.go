package server

import (
	"github.com/gofiber/fiber/v2"
)

type sessionsGetter interface {
	Get(c *fiber.Ctx) (sessionManager, error)
}

type sessionManager interface {
	sessionRecordGetter
	sessionRecordSetter
	sessionRecordDeleter
	sessionSaver
	sessionDestroyer
	sessionIdentifier
}

type sessionRecordSetterDeleterSaverDestroyer interface {
	sessionRecordSetter
	sessionRecordDeleter
	sessionSaver
	sessionDestroyer
}

type sessionRecordSetterDeleterSaver interface {
	sessionRecordSetter
	sessionRecordDeleter
	sessionSaver
}

type sessionRecordGetterSetterSaver interface {
	sessionRecordGetter
	sessionRecordSetter
	sessionSaver
}

type sessionRecordSetterSaver interface {
	sessionRecordSetter
	sessionSaver
}

type sessionRecordGetter interface {
	Get(key string) interface{}
}

type sessionRecordSetter interface {
	Set(key string, value interface{})
}

type sessionRecordDeleter interface {
	Delete(key string)
}

type sessionDestroyer interface {
	Destroy() error
}

type sessionSaver interface {
	Save() error
}

type sessionIdentifier interface {
	ID() string
}
