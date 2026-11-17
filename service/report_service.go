package service

import (
	"fmt"
	"math"
	"time"

	"manisora/model"
	"manisora/repository"
)

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func GetStockMutationReport(locationID int64, from, to string, itemID int64) (model.StockMutationReport, error) {
	report := model.StockMutationReport{
		LocationID: locationID,
		From:       from,
		To:         to,
		Items:      []model.StockMutation{},
	}

	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return report, fmt.Errorf("%w: from harus berformat YYYY-MM-DD", ErrInvalidInput)
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return report, fmt.Errorf("%w: to harus berformat YYYY-MM-DD", ErrInvalidInput)
	}
	if toDate.Before(fromDate) {
		return report, fmt.Errorf("%w: to tidak boleh sebelum from", ErrInvalidInput)
	}

	list, err := repository.GetStockMutation(locationID, from, to, itemID)
	if err != nil {
		return report, err
	}

	for i := range list {
		m := &list[i]
		m.Opening = round3(m.Opening)
		m.QtyIn = round3(m.QtyIn)
		m.QtyOut = round3(m.QtyOut)
		m.Closing = round3(m.Opening + m.QtyIn - m.QtyOut)
	}
	report.Items = list
	return report, nil
}
