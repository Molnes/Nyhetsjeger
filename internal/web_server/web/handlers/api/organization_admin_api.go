package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/access_control"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/access_settings_components"
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
	g.DELETE("/access-control/admin", oah.deleteAdminByEmail)
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

	return utils.Render(c, http.StatusCreated, access_settings_components.AdminTableRow(useradmin))
}

// expects email in json body
func (oah *OrganizationAdminApiHandler) deleteAdminByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing email")
	}

	err := access_control.RemoveAdminByEmail(oah.sharedData.DB, email)
	if err != nil {
		if err == access_control.ErrNoAdminWithGivenEmail {
			return echo.NewHTTPError(http.StatusNotFound, "No admin with given email found")
		}
		return err
	}

	return c.NoContent(http.StatusOK)
}
