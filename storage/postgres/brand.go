package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"market_system/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type brandRepo struct {
	db *pgxpool.Pool
}

func NewBrandRepo(db *pgxpool.Pool) *brandRepo {
	return &brandRepo{
		db: db,
	}
}

func (r *brandRepo) Create(ctx context.Context, req *models.CreateBrand) (*models.Brand, error) {

	var (
		branchId = uuid.New().String()
		query    = `
			INSERT INTO brand(
				id,
				name,
				updated_at
			) VALUES ($1, $2, NOW())
		`
	)

	_, err := r.db.Exec(ctx,
		query,
		branchId,
		req.Name,
	)

	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, &models.BrandPrimaryKey{Id: branchId})
}

func (r *brandRepo) GetByID(ctx context.Context, req *models.BrandPrimaryKey) (*models.Brand, error) {

	var (
		query = `
			SELECT
				 id,
				 name,
			FROM brand
			WHERE id = $1
		`
	)

	var (
		Id        sql.NullString
		Name      sql.NullString
		CreatedAt sql.NullString
		UpdatedAt sql.NullString
	)

	err := r.db.QueryRow(ctx, query, req.Id).Scan(
		&Id,
		&Name,
		&CreatedAt,
		&UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &models.Brand{
		Id:        Id.String,
		Name:      Name.String,
		CreatedAt: CreatedAt.String,
		UpdatedAt: UpdatedAt.String,
	}, nil
}
func (r *brandRepo) GetList(ctx context.Context, req *models.GetListBrandRequest) (*models.GetListBrandResponse, error) {
	var (
		resp   models.GetListBrandResponse
		where  = " WHERE TRUE"
		offset = " OFFSET 0"
		limit  = " LIMIT 10"
		sort   = " ORDER BY name DESC"
	)

	if req.Offset > 0 {
		offset = fmt.Sprintf(" OFFSET %d", req.Offset)
	}

	if req.Limit > 0 {
		limit = fmt.Sprintf(" LIMIT %d", req.Limit)
	}

	if len(req.Search) > 0 {
		where += " AND (name ILIKE '%" + req.Search + "%' OR phone ILIKE '%" + req.Search + "%')"
	}

	if len(req.Query) > 0 {
		where += req.Query
	}

	var query = `
		SELECT
			COUNT(*) OVER(),
			 id,
			 name,
			 created_at,
			 updated_at
		FROM brand
	`

	query += where + sort + offset + limit
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			Id        sql.NullString
			Name      sql.NullString
			CreatedAt sql.NullString
			UpdatedAt sql.NullString
		)

		err = rows.Scan(
			&resp.Count,
			&Id,
			&Name,
			&CreatedAt,
			&UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		resp.Brands = append(resp.Brands, &models.Brand{
			Id:        Id.String,
			Name:      Name.String,
			CreatedAt: CreatedAt.String,
			UpdatedAt: UpdatedAt.String,
		})

	}
	return &resp, nil
}

func (r *brandRepo) Update(ctx context.Context, req *models.UpdateBrand) (int64, error) {

	query := `
		UPDATE brand
			SET
				name = $2,
		WHERE id = $1
	`
	rowsAffected, err := r.db.Exec(ctx,
		query,
		req.Id,
		req.Name,
	)
	if err != nil {
		return 0, err
	}

	return rowsAffected.RowsAffected(), nil
}

func (r *brandRepo) Delete(ctx context.Context, req *models.BrandPrimaryKey) error {
	_, err := r.db.Exec(ctx, "DELETE FROM brand WHERE id = $1", req.Id)
	return err
}
