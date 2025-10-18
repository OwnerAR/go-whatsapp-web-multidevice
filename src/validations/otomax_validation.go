package validations

import (
	"context"
	"regexp"
	"strings"

	domainOtomax "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/otomax"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
)

// ValidateInsertInboxRequest validates InsertInbox request
func ValidateInsertInboxRequest(ctx context.Context, request domainOtomax.InsertInboxRequest) error {
	// Validate message content
	if strings.TrimSpace(request.Pesan) == "" {
		return pkgError.ValidationError("Message content (pesan) is required")
	}
	
	if len(request.Pesan) > 4096 {
		return pkgError.ValidationError("Message content is too long (max 4096 characters)")
	}
	
	// Validate reseller code
	if strings.TrimSpace(request.KodeReseller) == "" {
		return pkgError.ValidationError("Reseller code (kode_reseller) is required")
	}
	
	// Validate reseller code format (alphanumeric, underscore, hyphen allowed)
	resellerRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !resellerRegex.MatchString(request.KodeReseller) {
		return pkgError.ValidationError("Invalid reseller code format")
	}
	
	// Validate sender phone number
	if strings.TrimSpace(request.Pengirim) == "" {
		return pkgError.ValidationError("Sender phone number (pengirim) is required")
	}
	
	// Sanitize and validate phone number
	sanitizedPhone := utils.SanitizePhoneNumber(request.Pengirim)
	if !isValidPhoneNumber(sanitizedPhone) {
		return pkgError.ValidationError("Invalid phone number format")
	}
	
	// Validate sender type
	if strings.TrimSpace(request.TipePengirim) == "" {
		return pkgError.ValidationError("Sender type (tipe_pengirim) is required")
	}
	
	if request.TipePengirim != "W" {
		return pkgError.ValidationError("Sender type must be 'W' for WhatsApp")
	}
	
	return nil
}

// ValidateSetOutboxCallbackRequest validates SetOutboxCallback request
func ValidateSetOutboxCallbackRequest(ctx context.Context, request domainOtomax.SetOutboxCallbackRequest) error {
	// Validate URL
	if strings.TrimSpace(request.URL) == "" {
		return pkgError.ValidationError("Callback URL is required")
	}
	
	// Basic URL format validation
	if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
		return pkgError.ValidationError("Callback URL must start with http:// or https://")
	}
	
	if len(request.URL) > 2048 {
		return pkgError.ValidationError("Callback URL is too long (max 2048 characters)")
	}
	
	return nil
}

// ValidateTestRequest validates Test request
func ValidateTestRequest(ctx context.Context, request domainOtomax.TestRequest) error {
	// Validate phone number
	if strings.TrimSpace(request.Phone) == "" {
		return pkgError.ValidationError("Phone number is required")
	}
	
	// Sanitize and validate phone number
	sanitizedPhone := utils.SanitizePhoneNumber(request.Phone)
	if !isValidPhoneNumber(sanitizedPhone) {
		return pkgError.ValidationError("Invalid phone number format")
	}
	
	return nil
}

// ValidateGetRsRequest validates GetRs request
func ValidateGetRsRequest(ctx context.Context, request domainOtomax.GetRsRequest) error {
	// Validate reseller code
	if strings.TrimSpace(request.Kode) == "" {
		return pkgError.ValidationError("Reseller code is required")
	}
	
	// Validate reseller code format
	resellerRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !resellerRegex.MatchString(request.Kode) {
		return pkgError.ValidationError("Invalid reseller code format")
	}
	
	return nil
}

// ValidateGetSaldoRsRequest validates GetSaldoRs request
func ValidateGetSaldoRsRequest(ctx context.Context, request domainOtomax.GetSaldoRsRequest) error {
	// Validate reseller code
	if strings.TrimSpace(request.Kode) == "" {
		return pkgError.ValidationError("Reseller code is required")
	}
	
	// Validate reseller code format
	resellerRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !resellerRegex.MatchString(request.Kode) {
		return pkgError.ValidationError("Invalid reseller code format")
	}
	
	return nil
}

// ValidateCallbackPayload validates callback payload from OtomaX
func ValidateCallbackPayload(ctx context.Context, payload domainOtomax.CallbackPayload) error {
	// Validate transaction code
	if payload.Kode <= 0 {
		return pkgError.ValidationError("Invalid transaction code")
	}
	
	// Validate status code
	if payload.Status < 0 {
		return pkgError.ValidationError("Invalid status code")
	}
	
	// Validate message length if provided
	if payload.Message != "" && len(payload.Message) > 1024 {
		return pkgError.ValidationError("Message is too long (max 1024 characters)")
	}
	
	// Validate response message length if provided
	if payload.Pesan != "" && len(payload.Pesan) > 4096 {
		return pkgError.ValidationError("Response message is too long (max 4096 characters)")
	}
	
	// Validate sender phone number if provided
	if payload.Pengirim != "" {
		sanitizedPhone := utils.SanitizePhoneNumber(payload.Pengirim)
		if !isValidPhoneNumber(sanitizedPhone) {
			return pkgError.ValidationError("Invalid sender phone number format")
		}
	}
	
	return nil
}

// isValidPhoneNumber validates phone number format
func isValidPhoneNumber(phone string) bool {
	// Remove common prefixes and validate format
	phone = strings.TrimSpace(phone)
	
	// Remove common prefixes
	if strings.HasPrefix(phone, "+") {
		phone = phone[1:]
	}
	if strings.HasPrefix(phone, "62") {
		phone = "0" + phone[2:]
	}
	
	// Check if it's a valid Indonesian phone number format
	phoneRegex := regexp.MustCompile(`^08[0-9]{8,11}$`)
	return phoneRegex.MatchString(phone)
}

// SanitizeInsertInboxRequest sanitizes InsertInbox request data
func SanitizeInsertInboxRequest(request *domainOtomax.InsertInboxRequest) {
	request.Pesan = strings.TrimSpace(request.Pesan)
	request.KodeReseller = strings.TrimSpace(request.KodeReseller)
	request.Pengirim = utils.SanitizePhoneNumber(request.Pengirim)
	request.TipePengirim = strings.TrimSpace(strings.ToUpper(request.TipePengirim))
}

// SanitizeSetOutboxCallbackRequest sanitizes SetOutboxCallback request data
func SanitizeSetOutboxCallbackRequest(request *domainOtomax.SetOutboxCallbackRequest) {
	request.URL = strings.TrimSpace(request.URL)
}

// SanitizeTestRequest sanitizes Test request data
func SanitizeTestRequest(request *domainOtomax.TestRequest) {
	request.Phone = utils.SanitizePhoneNumber(request.Phone)
}

// SanitizeGetRsRequest sanitizes GetRs request data
func SanitizeGetRsRequest(request *domainOtomax.GetRsRequest) {
	request.Kode = strings.TrimSpace(request.Kode)
}

// SanitizeGetSaldoRsRequest sanitizes GetSaldoRs request data
func SanitizeGetSaldoRsRequest(request *domainOtomax.GetSaldoRsRequest) {
	request.Kode = strings.TrimSpace(request.Kode)
}

// SanitizeCallbackPayload sanitizes callback payload data
func SanitizeCallbackPayload(payload *domainOtomax.CallbackPayload) {
	payload.Message = strings.TrimSpace(payload.Message)
	payload.Pesan = strings.TrimSpace(payload.Pesan)
	if payload.Pengirim != "" {
		payload.Pengirim = utils.SanitizePhoneNumber(payload.Pengirim)
	}
}
