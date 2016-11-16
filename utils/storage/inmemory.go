package storage

type InmemoryStorage struct{}

func NewInmemoryStorage() *InmemoryStorage {
	return &InmemoryStorage{}
}

func (is *InmemoryStorage) Open(filename string, length int) ([]uint64, error) {
	return make([]uint64, length), nil
}

func (is *InmemoryStorage) Close() error {
	return nil
}

func (is *InmemoryStorage) Flush() error {
	return nil
}
