package otomax

// InsertInboxRequest represents the request structure for InsertInbox endpoint
type InsertInboxRequest struct {
	Pesan         string `json:"pesan" validate:"required"`         // Message content
	KodeReseller  string `json:"kode_reseller" validate:"omitempty"` // Reseller code
	Pengirim      string `json:"pengirim" validate:"required"`      // Sender phone number
	TipePengirim  string `json:"tipe_pengirim" validate:"required"` // Sender type (W for WhatsApp)
	KodeTerminal  int    `json:"kode_terminal" validate:"required"` // Terminal code
}

// InsertInboxResponse represents the response structure for InsertInbox endpoint
type InsertInboxResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		KodeInbox  int    `json:"kode_inbox"`  // Inbox code from OtomaX
		Status     int    `json:"status"`      // Status code from OtomaX
		StatusDesc string `json:"statusDesc"`  // Status description from OtomaX
		Pesan      string `json:"pesan"`       // Message content from OtomaX (for status 21)
	} `json:"result"`
}

// SetOutboxCallbackRequest represents the request structure for SetOutboxCallback endpoint
type SetOutboxCallbackRequest struct {
	URL string `json:"url" validate:"required,url"` // Callback URL for OtomaX responses
}

// SetOutboxCallbackResponse represents the response structure for SetOutboxCallback endpoint
type SetOutboxCallbackResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GetOutboxCallbackResponse represents the response structure for GetOutboxCallback endpoint
type GetOutboxCallbackResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		URL string `json:"url"` // Current callback URL
	} `json:"data"`
}

// CallbackPayload represents the payload structure that OtomaX sends to our callback URL
type CallbackPayload struct {
	Kode     int    `json:"kode"`               // Transaction code
	Status   int    `json:"status"`             // Status code
	Message  string `json:"message"`            // Response message
	Pesan    string `json:"pesan,omitempty"`    // Response content (if any)
	Pengirim string `json:"pengirim,omitempty"` // Original sender
}

// ResellerInfo represents reseller information structure
type ResellerInfo struct {
	Kode      string  `json:"kode"`      // Reseller code
	Nama      string  `json:"nama"`      // Reseller name
	Saldo     float64 `json:"saldo"`     // Current balance
	Status    string  `json:"status"`    // Reseller status
	IsActive  bool    `json:"is_active"` // Whether reseller is active
}

// GetRsRequest represents the request structure for GetRs endpoint
type GetRsRequest struct {
	Kode string `json:"kode" validate:"required"` // Reseller code
}

// GetRsResponse represents the response structure for GetRs endpoint
type GetRsResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    ResellerInfo  `json:"data"`
}

// GetSaldoRsRequest represents the request structure for GetSaldoRs endpoint
type GetSaldoRsRequest struct {
	Kode string `json:"kode" validate:"required"` // Reseller code
}

// GetSaldoRsResponse represents the response structure for GetSaldoRs endpoint
type GetSaldoRsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Kode string  `json:"kode"` // Reseller code
		Saldo float64 `json:"saldo"` // Current balance
	} `json:"data"`
}

// TestRequest represents the request structure for Test endpoint
type TestRequest struct {
	Phone string `json:"phone" validate:"required"` // Phone number for testing
}

// TestResponse represents the response structure for Test endpoint
type TestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Result string `json:"result"` // Test result
	} `json:"data"`
}

// GenericResponse represents a generic API response structure
type GenericResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
