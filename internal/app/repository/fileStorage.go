package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
)

type FileStore struct {
	fileName string
}

func NewFileStore(fileName string) (*FileStore, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	return &FileStore{
		fileName: fileName,
	}, nil
}

func (store *FileStore) SaveURL(ctx context.Context, u *URL) (*URL, error) {
	file, err := os.OpenFile(store.fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	u.UUID = string(fileInfo.Size())
	encoder := json.NewEncoder(file)
	encoder.Encode(u)

	return nil, nil
}

func (store *FileStore) LoadURL(_ context.Context, u *URL) (r *URL, err error) {
	file, err := os.OpenFile(store.fileName, os.O_RDONLY, 0666)

	if err != nil {
		return nil, newErrURLNotFound()
	}

	defer file.Close()

	urls := make([]URL, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var ur URL
		err := json.Unmarshal(scanner.Bytes(), &ur)

		if err != nil {
			continue
		}

		urls = append(urls, ur)
	}

	for _, v := range urls {
		if v.FullURL == u.FullURL {
			return &v, nil
		}

		if v.ShortURL == u.ShortURL {
			return &v, nil
		}
	}

	return nil, newErrURLNotFound()
}

func (store *FileStore) Ping(_ context.Context) error {
	return nil
}

func (store *FileStore) BatchURLS(ctx context.Context, urls []*URL) error {
	for _, u := range urls {
		store.SaveURL(ctx, u)
	}

	return nil
}

func (store *FileStore) GetByUID(ctx context.Context, id string) ([]*URL, error) {
	var urls []*URL

	return urls, nil
}
