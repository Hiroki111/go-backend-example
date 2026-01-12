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
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type ProductItem struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
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
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
}

type CreateProductResponse struct {
	Item ProductItem `json:"item"`
}

type UpdateProductRequest struct {
	Name       *string `json:"name"`
	PriceCents *int64  `json:"price_cents"`
}

type UpdateProductResponse struct {
	Item ProductItem `json:"item"`
}
type DeleteProductResponse struct {
	Message string `json:"message"`
}
