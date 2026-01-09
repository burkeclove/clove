
package requests

import (
	"mime/multipart"
)

type PutObject struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
	Name string `form:"name" binding:"required"`
}
