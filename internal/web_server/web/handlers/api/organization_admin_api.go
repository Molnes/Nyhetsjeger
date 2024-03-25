package api

import "github.com/Molnes/Nyhetsjeger/internal/config"

type OrganizationAdminApiHandler struct {
	sharedData *config.SharedData
}

// Creates a new AdminApiHandler
func NewOrganizationAdminApiHandler(sharedData *config.SharedData) *OrganizationAdminApiHandler {
	return &OrganizationAdminApiHandler{sharedData}
}
