package api

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/access_control"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/access_settings_components"
	"github.com/labstack/echo/v4"
)

const (
	hxReswap            = "HX-Reswap"
	hxOuterHTML         = "outerHTML"
	classPostAdminError = "post-admin-error"
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
		c.Response().Header().Set(hxReswap, hxOuterHTML)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(classPostAdminError, "Manglende epost"))
	}
	email = strings.ToLower(email)
	parsedAddress, err := mail.ParseAddress(email)
	if err != nil {
		c.Response().Header().Set(hxReswap, hxOuterHTML)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(classPostAdminError, "Feil epost formatering"))
	}
	if strings.Split(parsedAddress.Address, "@")[1] != "gmail.com" {
		c.Response().Header().Set(hxReswap, hxOuterHTML)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(classPostAdminError, "Kun Gmail addresser er støttet"))
	}

	session, err := oah.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		c.Response().Header().Set(hxReswap, hxOuterHTML)
		return err
	}
	caller := session.Values[sessions.USER_DATA_VALUE].(users.UserSessionData)
	if caller.Email == parsedAddress.Address {
		c.Response().Header().Set(hxReswap, hxOuterHTML)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(classPostAdminError, "Kan ikke redigere egen rolle"))
	}

	userAdmin, err := access_control.AddAdmin(oah.sharedData.DB, parsedAddress.Address)
	if err != nil {
		if err == access_control.ErrEmailAlreadyAdmin {
			c.Response().Header().Set(hxReswap, hxOuterHTML)
			return utils.Render(c, http.StatusConflict, components.ErrorText(classPostAdminError, "Gitte eposten allerede har rollen"))
		}
		return err
	}

	return utils.Render(c, http.StatusCreated, access_settings_components.AdminTableRow(userAdmin))
}

// Handles delete request to remove an admin by email. Email expected in query param 'email'.
func (oah *OrganizationAdminApiHandler) deleteAdminByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing email")
	}

	session, err := oah.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		return err
	}

	caller := session.Values[sessions.USER_DATA_VALUE].(users.UserSessionData)
	if caller.Email == email {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot modify own admin status")
	}

	err = access_control.RemoveAdminByEmail(oah.sharedData.DB, email)
	if err != nil {
		if err == access_control.ErrNoAdminWithGivenEmail {
			return echo.NewHTTPError(http.StatusNotFound, "No admin with given email found")
		}
		return err
	}
	// Returning no content with `200 OK` instead of `204 No Content`
	// due to HTMX refusing to replace content if response code is 204.
	// This is necessary to remove the row from the table.
	return c.NoContent(http.StatusOK)
}
