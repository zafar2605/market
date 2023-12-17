package helpers

import (
	"context"
	"market_system/config"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
)

type incrementRepo struct {
	db *pgxpool.Pool
}

func NewIncrementRepo(db *pgxpool.Pool) *incrementRepo {
	return &incrementRepo{
		db: db,
	}
}

func GetIncrementId(n int) string {
	t := "0000000"
	if len(strconv.Itoa(n+1)) == len(strconv.Itoa(n)) {
		return t[len(strconv.Itoa(n)):] + strconv.Itoa(n+1)
	}
	return t[len(strconv.Itoa(n))+1:] + strconv.Itoa(n+1)
}

func (i *incrementRepo) GetLast(tableName, columnName string) (string, error) {
	var last_Id string
	ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cancel()

	query := `SELECT ` + columnName + ` FROM ` + tableName + ` ORDER BY id DESC LIMIT 1`

	resp := i.db.QueryRow(ctx, query)

	if resp == nil {
		return "-0000001", nil
	}
	err := resp.Scan(
		&last_Id,
	)
	if err != nil {
		return "Can not scan from db", nil
	}
	number, _ := strconv.Atoi(last_Id[3:])
	incrementNumber := GetIncrementId(number)

	return incrementNumber, nil
}
