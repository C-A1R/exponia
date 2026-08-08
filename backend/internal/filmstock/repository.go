package filmstock

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/C-A1R/exponia/backend/internal/database/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("film stock not found")

type FilmFormat struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type FilmStock struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	Manufacturer string       `json:"manufacturer"`
	ISO          int32        `json:"iso"`
	ColorType    string       `json:"color_type"`
	Formats      []FilmFormat `json:"formats"`
	CreatedAt    time.Time    `json:"created_at"`
}

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
	}
}

func fromListRow(value db.ListFilmStocksRow) FilmStock {
	return FilmStock{
		ID:           value.ID,
		Manufacturer: value.Manufacturer,
		Name:         value.Name,
		ISO:          value.Iso,
		ColorType:    value.ColorType,
		CreatedAt:    value.CreatedAt.Time,
	}
}

func fromGetRow(value db.GetFilmStockByIDRow) FilmStock {
	return FilmStock{
		ID:           value.ID,
		Manufacturer: value.Manufacturer,
		Name:         value.Name,
		ISO:          value.Iso,
		ColorType:    value.ColorType,
		CreatedAt:    value.CreatedAt.Time,
	}
}

func (r *Repository) List(ctx context.Context) ([]FilmStock, error) {
	values, err := r.queries.ListFilmStocks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list film stocks: %w", err)
	}

	formatValues, err := r.queries.ListAllFilmStockFormats(ctx)
	if err != nil {
		return nil, fmt.Errorf("list film stock formats: %w", err)
	}

	formatsByStockID := make(map[int64][]FilmFormat)

	for _, value := range formatValues {
		formatsByStockID[value.FilmStockID] = append(
			formatsByStockID[value.FilmStockID],
			FilmFormat{
				ID:   value.ID,
				Code: value.Code,
				Name: value.Name,
			},
		)
	}

	stocks := make([]FilmStock, 0, len(values))

	for _, value := range values {
		stock := fromListRow(value)

		stock.Formats = formatsByStockID[stock.ID]

		if stock.Formats == nil {
			stock.Formats = []FilmFormat{}
		}

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (FilmStock, error) {
	value, err := r.queries.GetFilmStockByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FilmStock{}, ErrNotFound
		}

		return FilmStock{}, fmt.Errorf(
			"get film stock by id: %w",
			err,
		)
	}

	stock := fromGetRow(value)

	formatValues, err := r.queries.ListFilmStockFormats(ctx, id)
	if err != nil {
		return FilmStock{}, fmt.Errorf(
			"list formats for film stock %d: %w",
			id,
			err,
		)
	}

	stock.Formats = make([]FilmFormat, 0, len(formatValues))

	for _, value := range formatValues {
		stock.Formats = append(
			stock.Formats,
			FilmFormat{
				ID:   value.ID,
				Code: value.Code,
				Name: value.Name,
			},
		)
	}

	return stock, nil
}
