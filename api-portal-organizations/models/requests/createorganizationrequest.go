
package requests

type CreateOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
}
