package store

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"

	_ "github.com/go-sql-driver/mysql"

	"cart-service/model"
)

// MySQL implements Store using Amazon RDS (MySQL 8.0).
// The schema is created automatically on first connection.
type MySQL struct {
	db *sql.DB
}

// NewMySQL opens a connection pool and runs the schema migration.
// dsn format: "user:password@tcp(host:port)/dbname?parseTime=true"
func NewMySQL(dsn string) (*MySQL, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	m := &MySQL{db: db}
	if err := m.migrate(); err != nil {
		return nil, fmt.Errorf("mysql migrate: %w", err)
	}
	return m, nil
}

func (m *MySQL) migrate() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS carts (
			id          INT AUTO_INCREMENT PRIMARY KEY,
			customer_id INT         NOT NULL,
			status      VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at  TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS cart_items (
			id         INT AUTO_INCREMENT PRIMARY KEY,
			cart_id    INT NOT NULL,
			product_id INT NOT NULL,
			quantity   INT NOT NULL
		)`)
	return err
}

func (m *MySQL) CreateCart(ctx context.Context, customerID int) (int, error) {
	res, err := m.db.ExecContext(ctx,
		"INSERT INTO carts (customer_id) VALUES (?)", customerID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) GetCart(ctx context.Context, cartID int) (*model.Cart, []model.CartItem, error) {
	cart := &model.Cart{}
	err := m.db.QueryRowContext(ctx,
		"SELECT id, customer_id, status FROM carts WHERE id = ?", cartID).
		Scan(&cart.ID, &cart.CustomerID, &cart.Status)
	if err == sql.ErrNoRows {
		return nil, nil, &ErrNotFound{CartID: cartID}
	}
	if err != nil {
		return nil, nil, err
	}

	rows, err := m.db.QueryContext(ctx,
		"SELECT product_id, quantity FROM cart_items WHERE cart_id = ?", cartID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	return cart, items, rows.Err()
}

func (m *MySQL) AddItem(ctx context.Context, cartID, productID, quantity int) error {
	var count int
	if err := m.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM carts WHERE id = ?", cartID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return &ErrNotFound{CartID: cartID}
	}
	_, err := m.db.ExecContext(ctx,
		"INSERT INTO cart_items (cart_id, product_id, quantity) VALUES (?, ?, ?)",
		cartID, productID, quantity)
	return err
}

func (m *MySQL) Checkout(ctx context.Context, cartID int) (int, error) {
	res, err := m.db.ExecContext(ctx,
		"UPDATE carts SET status = 'checked_out' WHERE id = ? AND status = 'active'", cartID)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return 0, &ErrNotFound{CartID: cartID}
	}
	return rand.IntN(1_000_000) + 1, nil
}

func (m *MySQL) Close() { m.db.Close() }
