
package requests

type CreateApiKeyRequest struct {
	Name string `json:"name" binding:"required"`
}
