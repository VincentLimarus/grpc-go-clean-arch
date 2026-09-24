package gormrepo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"VincentLimarus/grpc-golang/internal/domain"
	"VincentLimarus/grpc-golang/internal/domain/entity"
	gormrepo "VincentLimarus/grpc-golang/internal/repository/gorm"
)

var productCols = []string{"id", "name", "price", "stock", "created_at", "updated_at"}

func setupRepo(t *testing.T) (*gormrepo.ProductRepository, sqlmock.Sqlmock, context.Context) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return gormrepo.NewProductRepository(gdb), mock, context.Background()
}

func newProduct() *entity.Product {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return &entity.Product{Name: "Laptop", Price: 1500, Stock: 4, CreatedAt: now, UpdatedAt: now}
}

func productRow(id int64, name string, price float64, stock int64) *sqlmock.Rows {
	return sqlmock.NewRows(productCols).AddRow(id, name, price, stock,
		time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC))
}

func TestCreateProduct_Success(t *testing.T) {
	repo, mock, ctx := setupRepo(t)
	p := newProduct()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "products"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))
	mock.ExpectCommit()

	require.NoError(t, repo.Create(ctx, p))
	assert.Equal(t, uint64(7), p.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct_DBError(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "products"`).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	err := repo.Create(ctx, newProduct())
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProduct_Success(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "products"`).
		WillReturnRows(productRow(1, "Laptop", 1500, 4))

	p, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, uint64(1), p.ID)
	assert.Equal(t, "Laptop", p.Name)
	assert.Equal(t, float64(1500), p.Price)
	assert.Equal(t, int32(4), p.Stock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProduct_NotFound(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "products"`).WillReturnRows(sqlmock.NewRows(productCols))

	p, err := repo.GetByID(ctx, 999)
	require.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProduct_DBError(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "products"`).WillReturnError(errors.New("db down"))

	p, err := repo.GetByID(ctx, 1)
	require.Error(t, err)
	assert.Nil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListProducts_Success(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "products"`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(2)))
	mock.ExpectQuery(`SELECT \* FROM "products" ORDER BY id`).
		WillReturnRows(productRow(1, "Laptop", 1500, 4).AddRow(int64(2), "Mouse", 25.0, int64(10),
			time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC),
			time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)))

	products, total, err := repo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, products, 2)
	assert.Equal(t, "Mouse", products[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListProducts_Pagination(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT count\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT \* FROM "products" ORDER BY id ASC LIMIT`).
		WillReturnRows(productRow(1, "Laptop", 1500, 4))

	products, total, err := repo.List(ctx, 2, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, products, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListProducts_Empty(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT count\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT \* FROM "products"`).WillReturnRows(sqlmock.NewRows(productCols))

	products, total, err := repo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, products)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListProducts_CountError(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT count\(\*\)`).WillReturnError(errors.New("db down"))

	products, total, err := repo.List(ctx, 1, 10)
	require.Error(t, err)
	assert.Nil(t, products)
	assert.Zero(t, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListProducts_FindError(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectQuery(`SELECT count\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT \* FROM "products"`).WillReturnError(errors.New("db down"))

	products, _, err := repo.List(ctx, 1, 10)
	require.Error(t, err)
	assert.Nil(t, products)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_Success(t *testing.T) {
	repo, mock, ctx := setupRepo(t)
	p := newProduct()
	p.ID = 1

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "products" SET`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Update(ctx, p))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo, mock, ctx := setupRepo(t)
	p := newProduct()
	p.ID = 1

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "products" SET`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.Update(ctx, p)
	require.ErrorIs(t, err, domain.ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_DBError(t *testing.T) {
	repo, mock, ctx := setupRepo(t)
	p := newProduct()
	p.ID = 1

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "products" SET`).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	require.Error(t, repo.Update(ctx, p))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_Success(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "products"`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Delete(ctx, 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_NotFound(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "products"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.Delete(ctx, 999)
	require.ErrorIs(t, err, domain.ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_DBError(t *testing.T) {
	repo, mock, ctx := setupRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "products"`).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	require.Error(t, repo.Delete(ctx, 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}
