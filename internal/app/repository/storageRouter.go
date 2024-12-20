package repository

type StorageRouter struct{}

func NewStorageRouter() *StorageRouter {
	return &StorageRouter{}
}

func (r *StorageRouter) GetStorage(fileName string) (Repository, error) {
	switch fileName {
	case "":
		return NewStore(), nil
	default:
		return NewFileStore(fileName)
	}
}
