package services

import (
	"log"
	"mime/multipart"
	"net/http"

	//"github.com/burkeclove/object-api/constants"
	"github.com/burkeclove/object-api/constants"
	"github.com/burkeclove/object-api/db"
	"github.com/burkeclove/object-api/models/requests"
	"github.com/gin-gonic/gin"
)

type ObjectService struct {
	MinioClient *db.MinioClient
}

func NewObjectService() *ObjectService {
	client := db.NewMinioClient()
	return &ObjectService{MinioClient: client}
}

func min(x, y int) int {
    if y < x {
        return y
    }
    return x
}

func (o *ObjectService) PutObject(c *gin.Context) {
	log.Println("Got a request to put an objecet yuh")
	var req requests.PutObject

	_ = c.ShouldBind(&req)
	
	// segment file	
	err := o.SegmentFile(req.File)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (o *ObjectService) SegmentFile(m *multipart.FileHeader) error {
	size := m.Size
	num_chunks := (size / constants.SegmentSize) + 1
	file := m.Open()
	for i := 0; i < num_chunks; i++ {
		start := i * constants.SegmentSize
		end := min(start + constants.SegmentSize, size)
		o.MinioClient.PutBytes(data []byte, name string)
	}
	return nil
}
