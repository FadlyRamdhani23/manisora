package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"manisora/database"
	"manisora/model"
	"manisora/repository"
)

func friendlyReceiptError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
		return errors.New("supplier, lokasi, atau barang tidak ditemukan")
	}
	return err
}

func CreateGoodsReceipt(gr model.GoodsReceipt) (model.GoodsReceipt, error) {
	if gr.SupplierID <= 0 || gr.LocationID <= 0 {
		return gr, errors.New("supplier_id dan location_id wajib diisi")
	}
	if len(gr.Items) == 0 {
		return gr, errors.New("minimal satu barang harus diisi")
	}
	for _, it := range gr.Items {
		if it.ItemID <= 0 || it.Qty <= 0 {
			return gr, errors.New("setiap barang harus punya item_id dan qty lebih dari 0")
		}
	}

	if gr.DocDate == "" {
		gr.DocDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", gr.DocDate); err != nil {
		return gr, errors.New("doc_date harus berformat YYYY-MM-DD")
	}
	gr.Status = "DRAFT"

	tx, err := database.DB.Begin()
	if err != nil {
		return gr, err
	}
	defer tx.Rollback()

	id, err := repository.CreateGoodsReceipt(tx, gr)
	if err != nil {
		return gr, friendlyReceiptError(err)
	}
	gr.ID = id
	gr.DocNo = fmt.Sprintf("GR-%s-%04d", gr.DocDate[:4], id)

	if err := repository.SetGoodsReceiptDocNo(tx, id, gr.DocNo); err != nil {
		return gr, err
	}

	for i := range gr.Items {
		itemID, err := repository.CreateGoodsReceiptItem(tx, id, gr.Items[i])
		if err != nil {
			return gr, friendlyReceiptError(err)
		}
		gr.Items[i].ID = itemID
		gr.Items[i].ReceiptID = id
	}

	if err := tx.Commit(); err != nil {
		return gr, err
	}
	return gr, nil
}
