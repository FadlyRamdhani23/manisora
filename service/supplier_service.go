package service

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"

	"manisora/model"
	"manisora/repository"
)

func GetAllSuppliers() ([]model.Supplier, error) {
	return repository.GetAllSuppliers()
}

func CreateSupplier(s model.Supplier) (model.Supplier, error) {
	s.Code = strings.TrimSpace(s.Code)
	s.Name = strings.TrimSpace(s.Name)

	if s.Code == "" || s.Name == "" {
		return s, errors.New("code dan name wajib diisi")
	}

	id, err := repository.CreateSupplier(s)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return s, errors.New("kode supplier sudah dipakai")
		}
		return s, err
	}
	s.ID = id
	return s, nil
}
