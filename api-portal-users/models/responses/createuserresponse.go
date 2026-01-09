
package responses

type CreateUserResponse struct {
	Id string `json:"id" binding:"required"`
	Jwt string `json:"jwt" binding:"required"`
}
