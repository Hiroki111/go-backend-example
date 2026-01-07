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
}

type LoginUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
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
