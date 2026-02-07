package event

// EventType constants to ensure consistency
const (
	TypePageViewed        = "page.viewed"
	TypeCTAClicked        = "cta.clicked"
	TypeCartItemAdded     = "cart.item_added"
	TypeCartItemRemoved   = "cart.item_removed"
	TypeCheckoutStarted   = "checkout.started"
	TypeCheckoutCompleted = "checkout.completed"
	TypePaymentProcessed  = "payment.processed"
	TypeOrderCreated      = "order.created"
	TypeConversion        = "conversion.completed" // "conversion.completed"
	TypeUserSignup        = "user.signup"
	TypeUserLogin         = "user.login"
	TypeUserLogout        = "user.logout"
)

// PageViewed event
type PageViewed struct {
	Metadata
	PageURL  string `json:"page_url"`
	Referrer string `json:"referrer,omitempty"`
}

// CTAClicked event
type CTAClicked struct {
	Metadata
	ButtonID    string `json:"button_id"`
	LinkURL     string `json:"link_url"`
	PageSection string `json:"page_section,omitempty"`
}

// CartItemAdded event
type CartItemAdded struct {
	Metadata
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
}

// CartItemRemoved event
type CartItemRemoved struct {
	Metadata
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CheckoutStarted event
type CheckoutStarted struct {
	Metadata
	CartID      string  `json:"cart_id"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`
	ItemCount   int     `json:"item_count"`
}

// CheckoutCompleted event
type CheckoutCompleted struct {
	Metadata
	OrderID     string  `json:"order_id"`
	CartID      string  `json:"cart_id"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`
}

// PaymentProcessed event
type PaymentProcessed struct {
	Metadata
	OrderID       string  `json:"order_id"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"` // success, failed
	PaymentMethod string  `json:"payment_method"`
}

// OrderCreated event
type OrderCreated struct {
	Metadata
	OrderID     string  `json:"order_id"`
	TotalAmount float64 `json:"total_amount"`
	ItemCount   int     `json:"item_count"`
}

// ConversionCompleted event
type ConversionCompleted struct {
	Metadata
	ConversionID string `json:"conversion_id"`
	Type         string `json:"type"` // e.g., "purchase", "signup"
	Value        string `json:"value,omitempty"`
}

// UserSignup event
type UserSignup struct {
	Metadata
	Email  string `json:"email"`
	Method string `json:"method"` // email, google, facebook
}

// UserLogin event
type UserLogin struct {
	Metadata
	Method string `json:"method"`
}

// UserLogout event
type UserLogout struct {
	Metadata
}
