package repository

type Repository interface {
	SaveURL(u *URL)
	LoadURL(u *URL) (r *URL, err error)
	IsUniqueShort(s string) bool
}

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
