package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/access_control"
	"github.com/labstack/echo/v4"
)

type OrganizationAdminApiHandler struct {
	sharedData *config.SharedData
}

// Creates a new OrganizationAdminApiHandler
func NewOrganizationAdminApiHandler(sharedData *config.SharedData) *OrganizationAdminApiHandler {
	return &OrganizationAdminApiHandler{sharedData}
}

// Registers the organization admin related handlers to the given echo group
func (oah *OrganizationAdminApiHandler) RegisterOrganizationAdminHandlers(g *echo.Group) {
	g.POST("/access-control/admin", oah.postAddAdminByEmail)
}

// Handles a post request to add an admin by email. Email expected in form data.
func (oah *OrganizationAdminApiHandler) postAddAdminByEmail(c echo.Context) error {
	email := c.FormValue("email")
	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing email")
	}

	useradmin, err := access_control.AddAdmin(oah.sharedData.DB, email)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, useradmin.Email)
}
