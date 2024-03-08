package utils

import (
	"context"

	"github.com/Molnes/Nyhetsjeger/internal/api/middlewares"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Returns the user id from the given echo context
// Must be used AFTER authentication middleware sets the user id in the context!!
func GetUserIDFromCtx(ctx echo.Context) uuid.UUID {
	return ctx.Get(middlewares.USER_ID_KEY).(uuid.UUID)
}

// Adds a value to context.Context of the request in the given echo.Context
// This key-value pair can be accessed in templates by `ctx.Value(key)`
func AddToContext(c echo.Context, key any, value any) {
	ctx := context.WithValue(c.Request().Context(), key, value)
	c.SetRequest(c.Request().WithContext(ctx))
}
