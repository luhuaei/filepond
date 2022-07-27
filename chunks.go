package filepond

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ChunkManager struct {
	filename  string
	totalSize uint64
	dir       string

	offset     uint64
	chunkFiles []string
}

func NewChunkManager(dir, filename string, totalSize uint64) *ChunkManager {
	return &ChunkManager{
		filename:  filename,
		dir:       dir,
		totalSize: totalSize,
	}
}

func (c *ChunkManager) Append(r io.ReadCloser) error {
	err := c.saveChunk(r)
	if err != nil {
		return err
	}

	if !c.Finish() {
		return nil
	}

	err = c.merge()
	if err != nil {
		return err
	}

	return c.clean()
}

func (c *ChunkManager) Offset() uint64 {
	return c.offset
}

func (c *ChunkManager) saveChunk(r io.ReadCloser) error {
	defer r.Close()

	name := filepath.Join(c.dir, fmt.Sprintf("%d.chunk", c.offset))
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	c.offset += uint64(n)
	if c.offset > c.totalSize {
		return errors.New("received data length more than totalsize")
	}

	c.chunkFiles = append(c.chunkFiles, name)
	return nil
}

func (c *ChunkManager) merge() error {
	f, err := os.Create(filepath.Join(c.dir, c.filename))
	if err != nil {
		return err
	}
	defer f.Close()

	var n int64
	for _, namePath := range c.chunkFiles {
		chunkFile, err := os.Open(namePath)
		if err != nil {
			return err
		}

		_n, err := io.Copy(f, chunkFile)
		chunkFile.Close() // close chunk file

		if err != nil {
			return err
		}

		n += _n
	}

	if uint64(n) != c.totalSize {
		return errors.New("merge file size incorrect")
	}

	return nil
}

func (c *ChunkManager) clean() error {
	for _, namePath := range c.chunkFiles {
		err := os.Remove(namePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ChunkManager) Finish() bool {
	return c.totalSize == c.offset
}
