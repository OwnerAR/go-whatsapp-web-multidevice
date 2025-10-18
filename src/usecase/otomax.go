package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainOtomax "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/otomax"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

type otomaxService struct {
	otomaxClient domainOtomax.IOtomaxClient
	sendService  domainSend.ISendUsecase
}

// NewOtomaxService creates a new OtomaX service
func NewOtomaxService(otomaxClient domainOtomax.IOtomaxClient, sendService domainSend.ISendUsecase) domainOtomax.IOtomaxUsecase {
	return &otomaxService{
		otomaxClient: otomaxClient,
		sendService:  sendService,
	}
}

// SendMessageToOtomax sends WhatsApp message to OtomaX via InsertInbox
func (s *otomaxService) SendMessageToOtomax(ctx context.Context, request domainOtomax.InsertInboxRequest) (*domainOtomax.InsertInboxResponse, error) {
	logrus.Infof("Sending message to OtomaX: %s from %s", request.Pesan, request.Pengirim)
	
	response, err := s.otomaxClient.InsertInbox(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to send message to OtomaX: %v", err)
		return nil, err
	}
	
	logrus.Infof("Successfully sent message to OtomaX, response: %+v", response)
	return response, nil
}

// SetCallbackURL configures the callback URL for OtomaX responses
func (s *otomaxService) SetCallbackURL(ctx context.Context, request domainOtomax.SetOutboxCallbackRequest) (*domainOtomax.SetOutboxCallbackResponse, error) {
	logrus.Infof("Setting OtomaX callback URL: %s", request.URL)
	
	response, err := s.otomaxClient.SetOutboxCallback(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to set OtomaX callback URL: %v", err)
		return nil, err
	}
	
	logrus.Infof("Successfully set OtomaX callback URL")
	return response, nil
}

// GetCallbackURL retrieves the current callback URL configuration
func (s *otomaxService) GetCallbackURL(ctx context.Context) (*domainOtomax.GetOutboxCallbackResponse, error) {
	response, err := s.otomaxClient.GetOutboxCallback(ctx)
	if err != nil {
		logrus.Errorf("Failed to get OtomaX callback URL: %v", err)
		return nil, err
	}
	
	return response, nil
}

// HandleOtomaxCallback processes callback responses from OtomaX
func (s *otomaxService) HandleOtomaxCallback(ctx context.Context, payload domainOtomax.CallbackPayload) error {
	logrus.Infof("Processing OtomaX callback: kode=%d, status=%d, message=%s", payload.Kode, payload.Status, payload.Message)
	
	// If there's a response message and sender, send it back to WhatsApp
	if payload.Pesan != "" && payload.Pengirim != "" {
		// Convert phone number to proper JID format
		phone := utils.SanitizePhoneNumber(payload.Pengirim)
		_, err := types.ParseJID(phone)
		if err != nil {
			logrus.Errorf("Failed to parse phone number %s: %v", phone, err)
			return err
		}
		
		// Create WhatsApp message request
		whatsappRequest := domainSend.MessageRequest{
			BaseRequest: domainSend.BaseRequest{
				Phone: phone,
			},
			Message: payload.Pesan,
		}
		
		// Send response back to WhatsApp
		_, err = s.sendService.SendText(ctx, whatsappRequest)
		if err != nil {
			logrus.Errorf("Failed to send OtomaX response to WhatsApp: %v", err)
			return err
		}
		
		logrus.Infof("Successfully sent OtomaX response to WhatsApp: %s", payload.Pesan)
	}
	
	return nil
}

// GetResellerInfo retrieves reseller information from OtomaX
func (s *otomaxService) GetResellerInfo(ctx context.Context, resellerCode string) (*domainOtomax.GetRsResponse, error) {
	request := domainOtomax.GetRsRequest{
		Kode: resellerCode,
	}
	
	response, err := s.otomaxClient.GetRs(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to get reseller info for %s: %v", resellerCode, err)
		return nil, err
	}
	
	return response, nil
}

// GetResellerBalance retrieves reseller balance from OtomaX
func (s *otomaxService) GetResellerBalance(ctx context.Context, resellerCode string) (*domainOtomax.GetSaldoRsResponse, error) {
	request := domainOtomax.GetSaldoRsRequest{
		Kode: resellerCode,
	}
	
	response, err := s.otomaxClient.GetSaldoRs(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to get reseller balance for %s: %v", resellerCode, err)
		return nil, err
	}
	
	return response, nil
}

// TestConnection tests the connection to OtomaX API
func (s *otomaxService) TestConnection(ctx context.Context, phoneNumber string) (*domainOtomax.TestResponse, error) {
	request := domainOtomax.TestRequest{
		Phone: phoneNumber,
	}
	
	response, err := s.otomaxClient.Test(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to test OtomaX connection: %v", err)
		return nil, err
	}
	
	return response, nil
}

// ValidateReseller validates if reseller exists and is active
func (s *otomaxService) ValidateReseller(ctx context.Context, resellerCode string) (bool, error) {
	resellerInfo, err := s.GetResellerInfo(ctx, resellerCode)
	if err != nil {
		return false, err
	}
	
	// Check if reseller exists and is active
	isValid := resellerInfo.Status == "success" && resellerInfo.Data.IsActive
	
	logrus.Debugf("Reseller %s validation result: %v", resellerCode, isValid)
	return isValid, nil
}

// ProcessWhatsAppMessage processes incoming WhatsApp message and forwards to OtomaX if needed
func (s *otomaxService) ProcessWhatsAppMessage(ctx context.Context, senderJID, messageText string) error {
	// Check if OtomaX integration is enabled
	if !config.OtomaxEnabled {
		logrus.Debugf("OtomaX integration is disabled, skipping message processing")
		return nil
	}
	
	// Check if we should forward this message type
	if !config.OtomaxForwardIncoming {
		logrus.Debugf("OtomaX forward incoming is disabled, skipping message processing")
		return nil
	}
	
	// Skip group messages if configured
	if !config.OtomaxForwardGroups && utils.IsGroupJID(senderJID) {
		logrus.Debugf("Skipping group message for OtomaX forwarding")
		return nil
	}
	
	// Extract phone number from JID
	phone := utils.ExtractPhoneFromJID(senderJID)
	if phone == "" {
		logrus.Errorf("Failed to extract phone number from JID: %s", senderJID)
		return fmt.Errorf("invalid sender JID: %s", senderJID)
	}
	
	// Create InsertInbox request
	request := domainOtomax.InsertInboxRequest{
		Pesan:        messageText,
		KodeReseller: config.OtomaxDefaultReseller,
		Pengirim:     phone,
		TipePengirim: "W", // W for WhatsApp
	}
	
	// Send to OtomaX
	response, err := s.SendMessageToOtomax(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to forward WhatsApp message to OtomaX: %v", err)
		return err
	}
	
	logrus.Infof("Successfully forwarded WhatsApp message to OtomaX: kode_inbox=%v, status=%v", response.Result.KodeInbox, response.Result.Status)
	
	// Check if status requires auto reply (41: Bukan Reseller, 42: Format Salah)
	if response.Result.Status == 41 || response.Result.Status == 42 {
		logrus.Infof("Status %d requires auto reply: %s", response.Result.Status, response.Result.StatusDesc)
		
		// Send auto reply message to WhatsApp
		err = s.sendAutoReplyToWhatsApp(ctx, senderJID, response.Result.StatusDesc)
		if err != nil {
			logrus.Errorf("Failed to send auto reply to WhatsApp: %v", err)
			// Don't return error here, just log it
		} else {
			logrus.Infof("Auto reply sent successfully to %s: %s", senderJID, response.Result.StatusDesc)
		}
	}
	
	return nil
}

// sendAutoReplyToWhatsApp sends auto reply message to WhatsApp user
func (s *otomaxService) sendAutoReplyToWhatsApp(ctx context.Context, senderJID, statusDesc string) error {
	// Extract phone number from JID
	phone := strings.Split(senderJID, "@")[0]
	
	// Create WhatsApp text message request
	whatsappRequest := domainSend.MessageRequest{
		BaseRequest: domainSend.BaseRequest{
			Phone: phone,
		},
		Message: statusDesc, // Use statusDesc as reply message
	}
	
	// Send text message using WhatsApp service
	_, err := s.sendService.SendText(ctx, whatsappRequest)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp text message: %w", err)
	}
	
	return nil
}

// SetupOtomaxCallback configures the callback URL for OtomaX responses
func (s *otomaxService) SetupOtomaxCallback(ctx context.Context, callbackURL string) error {
	if !config.OtomaxEnabled {
		logrus.Debugf("OtomaX integration is disabled, skipping callback setup")
		return nil
	}
	
	request := domainOtomax.SetOutboxCallbackRequest{
		URL: callbackURL,
	}
	
	_, err := s.SetCallbackURL(ctx, request)
	if err != nil {
		logrus.Errorf("Failed to setup OtomaX callback URL: %v", err)
		return err
	}
	
	logrus.Infof("Successfully setup OtomaX callback URL: %s", callbackURL)
	return nil
}

// ExtractResellerFromMessage extracts reseller code from message text
func (s *otomaxService) ExtractResellerFromMessage(messageText string) string {
	// Simple extraction logic - can be enhanced based on requirements
	// Example: "tiket.100000.1234" -> extract reseller from message format
	
	// For now, return default reseller
	// This can be enhanced to parse message format and extract reseller code
	return config.OtomaxDefaultReseller
}

// ShouldForwardMessage determines if a message should be forwarded to OtomaX
func (s *otomaxService) ShouldForwardMessage(ctx context.Context, senderJID, messageText string) bool {
	// Check basic conditions
	if !config.OtomaxEnabled || !config.OtomaxForwardIncoming {
		return false
	}
	
	// Skip empty messages
	if strings.TrimSpace(messageText) == "" {
		return false
	}
	
	// Skip group messages if not configured
	if !config.OtomaxForwardGroups && utils.IsGroupJID(senderJID) {
		return false
	}
	
	// Add more business logic here if needed
	// For example, check if message contains specific keywords or commands
	
	return true
}

