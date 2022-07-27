package main

import (
	"filepond"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	baseDir := filepath.Join("/tmp", "filepond-test")
	tempDir := filepath.Join(baseDir, "temp")
	saveDir := filepath.Join(baseDir, "save")

	fp, err := filepond.NewFilePond(tempDir, saveDir)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors)

	fp.Register(r.Group("/filepond"))

	d := &dummy{tempDir: tempDir, saveDir: saveDir}
	dg := r.Group("/dummy")
	dg.GET("tempFiles", d.tempFiles)
	dg.GET("saveFiles", d.saveFiles)

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
	c.Writer.Header().Add("Access-Control-Expose-Headers", "Upload-Offset, Content-Disposition")
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}
	c.Next()
}

type dummy struct {
	saveDir string
	tempDir string
}

// 返回 tempDir 中的 serverId
func (d *dummy) tempFiles(c *gin.Context) {
	d.readDir(d.tempDir, c)
}

func (d *dummy) saveFiles(c *gin.Context) {
	d.readDir(d.saveDir, c)
}

func (d *dummy) readDir(dir string, c *gin.Context) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			ids = append(ids, entry.Name())
		}
	}

	c.JSON(http.StatusOK, ids)
}
