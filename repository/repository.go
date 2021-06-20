package repository

import (
	"context"
	"encoding/json"

	"github.com/gosom/go-gdc/entities"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	iq      = `INSERT INTO individuals(registration_number, data) VALUES($1, $2)`
	searchQ = `SELECT data FROM individuals
				WHERE to_tsvector('english', data) @@ plainto_tsquery($1)`
)

type IndividualRepo struct {
	db *pgxpool.Pool
}

func NewIndividualRepo(db *pgxpool.Pool) (*IndividualRepo, error) {
	ans := IndividualRepo{
		db: db,
	}
	return &ans, nil
}

func (o *IndividualRepo) BulkInsert(ctx context.Context, items []entities.Individual) error {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return err
	}
	for i := range items {
		jd, err := json.Marshal(items[i])
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, iq, items[i].RegistrationNumber, jd); err != nil {
			return err
		}
	}
	defer tx.Rollback(ctx)
	return tx.Commit(ctx)
}

func (o *IndividualRepo) GetByRegNum(ctx context.Context, regNum string) (entities.Individual, error) {
	var item entities.Individual
	q := `select data from individuals WHERE registration_number = $1`
	var data []byte
	if err := o.db.QueryRow(ctx, q, regNum).Scan(&data); err != nil {
		return item, err
	}
	return item, json.Unmarshal(data, &item)
}

func (o *IndividualRepo) Select(ctx context.Context, conditions map[string]interface{}) ([]entities.Individual, error) {
	return nil, nil
}
func (o *IndividualRepo) Search(ctx context.Context, term string) ([]entities.Individual, error) {
	rows, err := o.db.Query(ctx, searchQ, term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entities.Individual
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var item entities.Individual
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
