package storage

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/edsrzf/mmap-go"
)

type PersistentStorage struct {
	file   *os.File
	mmaped mmap.MMap
}

func NewPersistentStorage() *PersistentStorage {
	return &PersistentStorage{}
}

func (ps *PersistentStorage) Open(filename string, length int) ([]uint64, error) {
	var err error
	ps.file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("error open file: %s", err.Error())
	}

	// seek to offset
	offset := int64(unsafe.Sizeof(uint64(0))) * int64(length)
	if _, err = ps.file.Seek(offset, 0); err != nil {
		return nil, ps.closeFileIfError("error seek file", err)
	}

	// write byte to change size of the file
	if _, err = ps.file.Write([]byte{0xFF}); err != nil {
		return nil, ps.closeFileIfError("error write file", err)
	}

	// map file in memory
	ps.mmaped, err = mmap.Map(ps.file, mmap.RDWR, 0)
	if err != nil {
		return nil, ps.closeFileIfError("error map file", err)
	}

	// cast mapped []byte to []uint64
	header := (*reflect.SliceHeader)(unsafe.Pointer(&ps.mmaped))
	header.Len, header.Cap = length, length

	return *(*[]uint64)(unsafe.Pointer(header)), nil
}

func (ps *PersistentStorage) closeFileIfError(msg string, rootErr error) error {
	if err := ps.file.Close(); err != nil {
		return fmt.Errorf("%s:%s. also error close file:%s", msg, rootErr.Error(), err.Error())
	}
	return fmt.Errorf("%s:%s", msg, rootErr.Error())
}

func (ps *PersistentStorage) Close() error {
	if err := ps.mmaped.Flush(); err != nil {
		return err
	}

	if err := ps.mmaped.Unmap(); err != nil {
		return err
	}

	if err := ps.file.Close(); err != nil {
		return err
	}

	return nil
}

func (ps *PersistentStorage) Flush() error {
	return ps.mmaped.Flush()
}
