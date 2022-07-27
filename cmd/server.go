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

	fp.Register(r.Group("/filepond"))

	err = r.Run("127.0.0.1:8888")
	if err != nil {
		panic(err)
	}
}

func cors(c *gin.Context) {
	c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Add("Access-Control-Allow-Headers", "*")
	c.Writer.Header().Add("Access-Control-Allow-Methods", "*")
	c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Add("Access-Control-Expose-Headers", "Upload-Offset")
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}
	c.Next()
}
