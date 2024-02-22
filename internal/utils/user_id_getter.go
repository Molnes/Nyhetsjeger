package utils

import (
	"github.com/Molnes/Nyhetsjeger/internal/api/middlewares"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Returns the user id from the given echo context
// Must be used AFTER authentication middleware sets the user id in the context!!
func GetUserIDFromCtx(ctx echo.Context) uuid.UUID {
	return ctx.Get(middlewares.USER_ID_KEY).(uuid.UUID)
}
