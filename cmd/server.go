package main

import (
	"filepond"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	baseDir := filepath.Join("/tmp", "filepond-test")
	fp, err := filepond.NewFilePond(filepath.Join(baseDir, "temp"), filepath.Join(baseDir, "save"))
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors)

	g := r.Group("")
	g.POST("/filepond", fp.Process)
	g.DELETE("/filepond", fp.Revert)

	err = r.Run("127.0.0.1:8888")
	if err != nil {
		panic(err)
	}
}

func cors(c *gin.Context) {
	c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Add("Access-Control-Allow-Headers", "Accept, X-Access-Token, X-Application-Name, X-Request-Sent-Time")
	c.Writer.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
	c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}
	c.Next()
}
