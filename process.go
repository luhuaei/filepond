// api docment: https://pqina.nl/filepond/docs/api/server
package filepond

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FilePond 中的 api 接口分为两类
// - 一种是表单没有提交之前的，上传的文件会保留在 tempDir 中的。
// - 一种是表单已经提交，上传的资源已经从 tempDir 目录移动到最终的位置 saveDir 。
type FilePond struct {
	tempDir string
	saveDir string

	uploading sync.Map // map[id]ChunkManager
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

func (fp *FilePond) Register(g *gin.RouterGroup) {
	g.POST("", fp.Post)
	g.GET("", fp.Get)
	g.DELETE("", fp.Revert)
	g.PATCH("", fp.ProcessChunksPatch)
	g.HEAD("", fp.ProcessChunksHead)
}

func (fp *FilePond) Post(c *gin.Context) {
	length := c.GetHeader("Upload-Length")
	if length == "" {
		fp.ProcessNoChunks(c)
	} else {
		fp.ProcessChunks(c)
	}
}

func (fp *FilePond) Get(c *gin.Context) {
	if c.Query("restore") != "" {
		fp.Restore(c)
	} else if c.Query("load") != "" {
		fp.Load(c)
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

func (fp *FilePond) Head(c *gin.Context) {
	if c.Query("patch") != "" {
		fp.ProcessChunksHead(c)
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

// 1. FilePond uploads file my-file.jpg as multipart/form-data using a POST request
// 2. server saves file to unique location tmp/12345/my-file.jpg
// 3. server returns unique location id 12345 in text/plain response
// 4. FilePond stores unique id 12345 in a hidden input field
// 5. client submits the FilePond parent form containing the hidden input field with the unique id
// 6. server uses the unique id to move tmp/12345/my-file.jpg to its final location and remove the tmp/12345 folder
func (fp *FilePond) ProcessNoChunks(c *gin.Context) {
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

// 1. FilePond requests a transfer id from the server, a unique location to identify this transfer with. It does this using a POST request. The request is accompanied by the metadata and the total file upload size set to the Upload-Length header.
// 2. server create unique location tmp/12345/
// 3. server returns unique location id 12345 in text/plain response
// 4. FilePond stores unique id 12345 in file item
// 5. FilePond sends first chunk using a PATCH request adding the unique id 12345 in the URL, each PATCH request is accompanied by a Upload-Offset, Upload-Length, and Upload-Name header. The Upload-Offset header contains the byte offset of the chunk, the Upload-Length header contains the total file size, the Upload-Name header contains the file name.
// 6. FilePond sends chunks until all chunks have been uploaded succesfully.
// 7. server creates the file if all chunks have been received succesfully.
// 8. FilePond stores the unique id 12345 as the server id of this file.
// 9. client submits the FilePond parent form containing the hidden input field with the unique id
// 10. server uses the unique id to move tmp/12345/my-file.jpg to its final location and remove the tmp/12345 folder
//
// if one of the chunks fails to upload after the set amount of retries in chunkRetryDelays the user has the option to retry the upload.
// FilePond As FilePond remembers the previous transfer id the process now starts of with a HEAD request accompanied by the transfer id (12345) in the URL.
// server responds with Upload-Offset set to the next expected chunk offset in bytes.
// FilePond marks all chunks with lower offsets as complete and continues with uploading the chunk at the requested offset.

func (fp *FilePond) ProcessChunks(c *gin.Context) {
	id := uuid.NewString()
	err := os.Mkdir(filepath.Join(fp.tempDir, id), 0755)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, id)
}

func (fp *FilePond) ProcessChunksPatch(c *gin.Context) {
	id := c.Query("patch")
	if id == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("not found patch id"))
		return
	}

	var manager *ChunkManager
	var err error
	_m, ok := fp.uploading.Load(id)
	if ok {
		manager = _m.(*ChunkManager)
	} else {
		manager, err = fp.initChunkManager(id, c)
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = manager.Append(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if manager.Finish() {
		fp.uploading.Delete(id)
	}

	c.Status(http.StatusOK)
}

func (fp *FilePond) initChunkManager(id string, c *gin.Context) (*ChunkManager, error) {
	filename := c.GetHeader("Upload-Name")
	_totalSize := c.GetHeader("Upload-Length")
	totalSize, err := strconv.Atoi(_totalSize)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(fp.tempDir, id)
	m := &ChunkManager{
		filename:  filename,
		dir:       dir,
		totalSize: uint64(totalSize),
	}

	fp.uploading.Store(id, m)

	return m, nil
}

func (fp *FilePond) ProcessChunksHead(c *gin.Context) {
	id := c.Query("patch")
	if id == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("not found patch id"))
		return
	}

	_m, ok := fp.uploading.Load(id)
	if !ok {
		c.AbortWithError(http.StatusInternalServerError, errors.New("patch id not uploading file"))
		return
	}

	manager := _m.(*ChunkManager)
	c.Writer.Header().Set("Upload-Offset", fmt.Sprintf("%d", manager.Offset()))
	c.Status(http.StatusOK)
}

// Restore 是表单没有提交，但上传已经完成，资源存储在 tempDir 中
// 1. FilePond requests restore of file with id 12345 using a GET request
// 2. server returns a file object with header Content-Disposition: inline; filename="my-file.jpg"
func (fp *FilePond) Restore(c *gin.Context) {}

// Load 是表单已经提交，资源被移动到最终的存储位置
// 1. FilePond requests restore of file with id 12345 or a file name using a GET request
// 2. server returns a file object with header Content-Disposition: inline; filename="my-file.jpg"
func (fp *FilePond) Load(c *gin.Context) {}

// Fetch 获取资源
func (fp *FilePond) Fetch(c *gin.Context) {}

// Remove 为一个自定义的函数，参数不固定
func (fp *FilePond) Remove(c *gin.Context) {}
