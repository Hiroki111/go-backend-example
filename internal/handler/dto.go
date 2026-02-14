package handler

type ErrorResponse struct {
	Error string `json:"error"`
}

type RegisterUserResponse struct {
	Status string `json:"status"`
}

type LoginUserResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type RegisterUserRequest struct {
	UserName string `json:"user_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=2,max=100"`
}

type LoginUserRequest struct {
	UserName string `json:"user_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=2,max=100"`
}

type ProductItem struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	PriceCents uint   `json:"price_cents"`
}

type GetProductsResponse struct {
	Items   []ProductItem `json:"items"`
	Page    int           `json:"page"`
	Limit   int           `json:"limit"`
	Total   int           `json:"total"`
	HasNext bool          `json:"hasNext"`
}

type GetProductResponse struct {
	Item ProductItem `json:"item"`
}

type CreateProductRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=100"`
	PriceCents uint   `json:"price_cents"`
}

type CreateProductResponse struct {
	Item ProductItem `json:"item"`
}

type UpdateProductRequest struct {
	Name       *string `json:"name" validate:"min=2,max=100"`
	PriceCents *uint   `json:"price_cents"`
}

type UpdateProductResponse struct {
	Item ProductItem `json:"item"`
}
type DeleteProductResponse struct {
	Message string `json:"message"`
}

type CreateOrderRequest struct {
	ProductID uint `json:"product_id"`
}

type CreateOrderResponse struct {
	Message string `json:"message"`
}

type OrderItem struct {
	ID           uint   `json:"id"`
	CustomerName string `json:"customer_name"`
	ProductName  string `json:"product_name"`
	PriceCents   uint   `json:"price_cents"`
}

type GetOrdersResponse struct {
	Items   []OrderItem `json:"items"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Total   int         `json:"total"`
	HasNext bool        `json:"hasNext"`
}

type GetOrderResponse struct {
	Item OrderItem `json:"item"`
}

type UpdateOrderResponse struct {
	Item OrderItem `json:"item"`
}

type UpdateOrderRequest struct {
	PriceCents *uint `json:"price_cents"`
}

type DeleteOrderResponse struct {
	Message string `json:"message"`
}
