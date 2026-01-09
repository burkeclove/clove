package main

import (
	"os"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/burkeclove/object-api/services"
)

func main() {

	log.SetOutput(os.Stdout)
	log.Println("BOOT")

	r := gin.Default()

	// create handler
	service := services.NewObjectService()

	obj := r.Group("/api/objects")

	obj.POST("", service.PutObject)

	log.Println("About to serve on :8080")
    r.Run(":8080")
}



