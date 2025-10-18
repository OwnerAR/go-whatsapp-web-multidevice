package otomax

import "context"

// IOtomaxUsecase defines the interface for OtomaX integration use cases
type IOtomaxUsecase interface {
	// SendMessageToOtomax sends WhatsApp message to OtomaX via InsertInbox
	SendMessageToOtomax(ctx context.Context, request InsertInboxRequest) (*InsertInboxResponse, error)
	
	// SetCallbackURL configures the callback URL for OtomaX responses
	SetCallbackURL(ctx context.Context, request SetOutboxCallbackRequest) (*SetOutboxCallbackResponse, error)
	
	// GetCallbackURL retrieves the current callback URL configuration
	GetCallbackURL(ctx context.Context) (*GetOutboxCallbackResponse, error)
	
	// HandleOtomaxCallback processes callback responses from OtomaX
	HandleOtomaxCallback(ctx context.Context, payload CallbackPayload) error
	
	// GetResellerInfo retrieves reseller information from OtomaX
	GetResellerInfo(ctx context.Context, resellerCode string) (*GetRsResponse, error)
	
	// GetResellerBalance retrieves reseller balance from OtomaX
	GetResellerBalance(ctx context.Context, resellerCode string) (*GetSaldoRsResponse, error)
	
	// TestConnection tests the connection to OtomaX API
	TestConnection(ctx context.Context, phoneNumber string) (*TestResponse, error)
	
	// ValidateReseller validates if reseller exists and is active
	ValidateReseller(ctx context.Context, resellerCode string) (bool, error)
	
	// ProcessWhatsAppMessage processes incoming WhatsApp message and forwards to OtomaX if needed
	ProcessWhatsAppMessage(ctx context.Context, senderJID, messageText string) error
}

// IOtomaxClient defines the interface for OtomaX API client
type IOtomaxClient interface {
	// InsertInbox sends message to OtomaX inbox
	InsertInbox(ctx context.Context, request InsertInboxRequest) (*InsertInboxResponse, error)
	
	// SetOutboxCallback sets the callback URL for outbox responses
	SetOutboxCallback(ctx context.Context, request SetOutboxCallbackRequest) (*SetOutboxCallbackResponse, error)
	
	// GetOutboxCallback gets the current callback URL
	GetOutboxCallback(ctx context.Context) (*GetOutboxCallbackResponse, error)
	
	// GetRs gets reseller information
	GetRs(ctx context.Context, request GetRsRequest) (*GetRsResponse, error)
	
	// GetSaldoRs gets reseller balance
	GetSaldoRs(ctx context.Context, request GetSaldoRsRequest) (*GetSaldoRsResponse, error)
	
	// Test tests the API connection
	Test(ctx context.Context, request TestRequest) (*TestResponse, error)
	
	// GenerateAuthToken generates authentication token for API requests
	GenerateAuthToken(requestBody string) (string, error)
}
