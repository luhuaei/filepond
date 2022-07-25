package filepond

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FilePond struct {
	tempDir string
	saveDir string
}

func NewFilePond(tempDir, saveDir string) (*FilePond, error) {
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(saveDir, 0755)
	if err != nil {
		return nil, err
	}

	return &FilePond{
		tempDir: tempDir,
		saveDir: saveDir,
	}, nil
}

// 1. FilePond uploads file my-file.jpg as multipart/form-data using a POST request
// 2. server saves file to unique location tmp/12345/my-file.jpg
// 3. server returns unique location id 12345 in text/plain response
// 4. FilePond stores unique id 12345 in a hidden input field
// 5. client submits the FilePond parent form containing the hidden input field with the unique id
// 6. server uses the unique id to move tmp/12345/my-file.jpg to its final location and remove the tmp/12345 folder
func (fp *FilePond) Process(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("not 'file' partform data"))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer f.Close()

	id := uuid.NewString()
	err = os.Mkdir(filepath.Join(fp.tempDir, id), 0755)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	dst, err := os.Create(filepath.Join(fp.tempDir, id, file.Filename))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer dst.Close()
	io.Copy(dst, f)

	c.JSON(http.StatusOK, id)
}

// FilePond sends DELETE request with 12345 as body by tapping the undo button
// server removes temporary folder matching the supplied id tmp/12345 and returns an empty response
func (fp *FilePond) Revert(c *gin.Context) {
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusNoContent, errors.New("body not content"))
		return
	}
	defer c.Request.Body.Close()

	id, err := uuid.ParseBytes(b)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("body not id"))
		return
	}

	err = os.RemoveAll(filepath.Join(fp.tempDir, id.String()))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
