package service

import (
	"ya_url_shortener/internal/repository"
)

func CreateResource(storage repository.MemoryStorage, url string) string {
	newResource := repository.CreateResource(storage, url)
	shortenedUrl := encodeUrl(newResource.Identifier)
	newResource.Shortened = shortenedUrl
	return shortenedUrl
}

func GetResource(storage repository.MemoryStorage, id int32) (string, error) {
	r, err := repository.GetResource(storage, id)
	if err != nil {
		return "", err
	}
	return r.Address, nil
}
