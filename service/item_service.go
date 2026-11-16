package service

import (
	"errors"
	"manisora/model"
	"manisora/repository"
	"strings"
)

func GetAllItems() ([]model.Item, error) {
	return repository.GetAllItems()
}

func CreateItem(item model.Item) (model.Item, error) {
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.BaseUnit = strings.TrimSpace(item.BaseUnit)

	if item.Code == "" || item.Name == "" || item.BaseUnit == "" {
		return item, errors.New("code, name, dan base_unit wajib diisi")
	}

	id, err := repository.CreateItem(item)
	if err != nil {
		return item, err
	}
	item.ID = id
	return item, nil
}
