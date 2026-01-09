
package responses

type CreateOrganizationResponse struct {
	Name string `json:"name" binding:"required"`
	Id string `json:"id" binding:"required"`
	ApiKeyId string `json:"api_key_id" binding:"required"`
	ApiKey string `json:"api_key" binding:"required"`
}
