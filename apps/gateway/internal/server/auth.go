package server

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// apiKey guards a route group with a shared key via X-API-Key or Bearer token,
// compared in constant time.
func apiKey(key string) echo.MiddlewareFunc {
	want := []byte(key)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			got := c.Request().Header.Get("X-API-Key")
			if got == "" {
				if b := c.Request().Header.Get("Authorization"); strings.HasPrefix(b, "Bearer ") {
					got = strings.TrimPrefix(b, "Bearer ")
				}
			}
			if subtle.ConstantTimeCompare([]byte(got), want) != 1 {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			return next(c)
		}
	}
}
