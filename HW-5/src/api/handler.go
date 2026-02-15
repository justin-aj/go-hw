package api

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// ProductServer implements ServerInterface with in-memory storage
type ProductServer struct {
	products sync.Map
}

// NewProductServer creates a new server with initialized storage
func NewProductServer() *ProductServer {
	return &ProductServer{}
}

// ============================================================
// PRODUCT ENDPOINTS
// ============================================================

// GetProduct - GET /products/{productId}
func (s *ProductServer) GetProduct(ctx echo.Context, productId int32) error {
	if productId < 1 {
		return ctx.JSON(http.StatusBadRequest, Error{
			Error:   "INVALID_INPUT",
			Message: "Product ID must be a positive integer",
		})
	}

	val, exists := s.products.Load(productId)
	if !exists {
		return ctx.JSON(http.StatusNotFound, Error{
			Error:   "NOT_FOUND",
			Message: "Product not found",
		})
	}

	product := val.(Product)
	return ctx.JSON(http.StatusOK, product)
}

// AddProductDetails - POST /products/{productId}/details
func (s *ProductServer) AddProductDetails(ctx echo.Context, productId int32) error {
	if productId < 1 {
		return ctx.JSON(http.StatusBadRequest, Error{
			Error:   "INVALID_INPUT",
			Message: "Product ID must be a positive integer",
		})
	}

	var product Product
	if err := ctx.Bind(&product); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{
			Error:   "INVALID_INPUT",
			Message: "Invalid JSON in request body",
		})
	}

	if err := validateProduct(product); err != nil {
		return ctx.JSON(http.StatusBadRequest, *err)
	}

	if product.ProductId != productId {
		detail := "Product ID in body does not match URL path"
		return ctx.JSON(http.StatusBadRequest, Error{
			Error:   "INVALID_INPUT",
			Message: "Product ID mismatch",
			Details: &detail,
		})
	}

	s.products.Store(productId, product)

	return ctx.NoContent(http.StatusNoContent)
}

// validateProduct checks all constraints from the YAML spec
func validateProduct(p Product) *Error {
	if p.ProductId < 1 {
		detail := "product_id must be >= 1"
		return &Error{Error: "INVALID_INPUT", Message: "Invalid product_id", Details: &detail}
	}

	if len(p.Sku) == 0 {
		detail := "sku is required"
		return &Error{Error: "INVALID_INPUT", Message: "Missing sku", Details: &detail}
	}
	if len(p.Sku) > 100 {
		detail := "sku must be at most 100 characters"
		return &Error{Error: "INVALID_INPUT", Message: "Invalid sku", Details: &detail}
	}

	if len(p.Manufacturer) == 0 {
		detail := "manufacturer is required"
		return &Error{Error: "INVALID_INPUT", Message: "Missing manufacturer", Details: &detail}
	}
	if len(p.Manufacturer) > 200 {
		detail := "manufacturer must be at most 200 characters"
		return &Error{Error: "INVALID_INPUT", Message: "Invalid manufacturer", Details: &detail}
	}

	if p.CategoryId < 1 {
		detail := "category_id must be >= 1"
		return &Error{Error: "INVALID_INPUT", Message: "Invalid category_id", Details: &detail}
	}

	if p.Weight < 0 {
		detail := "weight must be >= 0"
		return &Error{Error: "INVALID_INPUT", Message: "Invalid weight", Details: &detail}
	}

	if p.SomeOtherId < 1 {
		detail := "some_other_id must be >= 1"
		return &Error{Error: "INVALID_INPUT", Message: "Invalid some_other_id", Details: &detail}
	}

	return nil
}

// ============================================================
// STUB ENDPOINTS (not required for this assignment)
// ============================================================

func (s *ProductServer) ProcessPayment(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, Error{
		Error:   "NOT_IMPLEMENTED",
		Message: "Payment processing not implemented",
	})
}

func (s *ProductServer) CreateShoppingCart(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, Error{
		Error:   "NOT_IMPLEMENTED",
		Message: "Shopping cart not implemented",
	})
}

func (s *ProductServer) CheckoutCart(ctx echo.Context, shoppingCartId int32) error {
	return ctx.JSON(http.StatusNotImplemented, Error{
		Error:   "NOT_IMPLEMENTED",
		Message: "Cart checkout not implemented",
	})
}

func (s *ProductServer) AddItemsToCart(ctx echo.Context, shoppingCartId int32) error {
	return ctx.JSON(http.StatusNotImplemented, Error{
		Error:   "NOT_IMPLEMENTED",
		Message: "Add items to cart not implemented",
	})
}

func (s *ProductServer) ReserveInventory(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, Error{
		Error:   "NOT_IMPLEMENTED",
		Message: "Reserve inventory not implemented",
	})
}

func (s *ProductServer) ShipProduct(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, Error{
		Error:   "NOT_IMPLEMENTED",
		Message: "Ship product not implemented",
	})
}