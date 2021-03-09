package server

// formValueGetter is a fiber.Ctx abstraction.
type formValueGetter interface {
	FormValue(key string, defaultValue ...string) string
}
