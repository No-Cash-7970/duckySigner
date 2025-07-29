package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// RootGet is the route handler for `GET /`
func RootGet(wailsApp *application.App) func(echo.Context) error {
	return func(c echo.Context) error {
		wailsApp.Event.Emit("session_init_response", []string{"account 1", "account 2"})
		return c.JSON(http.StatusOK, "OK")
	}
}
