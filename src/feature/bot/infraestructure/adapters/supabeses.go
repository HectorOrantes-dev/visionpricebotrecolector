package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type SupabaseRepositoryAdapter struct {
	db *sql.DB
}

func NewSupabaseRepositoryAdapter(db *sql.DB) *SupabaseRepositoryAdapter {
	return &SupabaseRepositoryAdapter{
		db: db,
	}
}

func (r *SupabaseRepositoryAdapter) Upsert(ctx context.Context, product *entities.Product) error {
	query := `
		INSERT INTO products (id, ml_id, nombre, descripcion, precio, moneda, categoria, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (ml_id) DO UPDATE
		SET nombre = EXCLUDED.nombre,
			descripcion = COALESCE(NULLIF(EXCLUDED.descripcion, ''), products.descripcion),
			precio = EXCLUDED.precio,
			moneda = EXCLUDED.moneda,
			categoria = EXCLUDED.categoria
		RETURNING id;
	`
	err := r.db.QueryRowContext(ctx, query,
		product.ID,
		product.MLID,
		product.Nombre,
		product.Descripcion,
		product.Precio,
		product.Moneda,
		product.Categoria,
		product.CreatedAt,
	).Scan(&product.ID)

	if err != nil {
		return fmt.Errorf("error upserting product to database: %w", err)
	}

	return nil
}

func (r *SupabaseRepositoryAdapter) SaveSnapshot(ctx context.Context, snapshot *entities.PriceSnapshot) error {
	query := `
		INSERT INTO price_snapshots (id, product_id, precio, moneda, fetched_at)
		VALUES ($1, $2, $3, $4, $5);
	`
	_, err := r.db.ExecContext(ctx, query,
		snapshot.ID,
		snapshot.ProductID,
		snapshot.Precio,
		snapshot.Moneda,
		snapshot.FetchedAt,
	)
	if err != nil {
		return fmt.Errorf("error inserting price snapshot to database: %w", err)
	}

	return nil
}

func (r *SupabaseRepositoryAdapter) ListByCategory(ctx context.Context, category string) ([]entities.Product, error) {
	query := `
		SELECT id, ml_id, nombre, descripcion, precio, moneda, categoria, created_at
		FROM products
		WHERE categoria = $1;
	`
	rows, err := r.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("error querying products by category: %w", err)
	}
	defer rows.Close()

	var products []entities.Product
	for rows.Next() {
		var p entities.Product
		err := rows.Scan(
			&p.ID,
			&p.MLID,
			&p.Nombre,
			&p.Descripcion,
			&p.Precio,
			&p.Moneda,
			&p.Categoria,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during products iteration: %w", err)
	}

	return products, nil
}
