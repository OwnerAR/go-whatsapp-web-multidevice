package otomax

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainOtomax "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/otomax"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/sirupsen/logrus"
)

type Client struct {
	baseURL    string
	appID      string
	appKey     string
	devKey     string
	httpClient *http.Client
}

// NewOtomaxClient creates a new OtomaX API client
func NewOtomaxClient() domainOtomax.IOtomaxClient {
	return &Client{
		baseURL: config.OtomaxAPIURL,
		appID:   config.OtomaxAppID,
		appKey:  config.OtomaxAppKey,
		devKey:  config.OtomaxDevKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateAuthToken generates HMAC-SHA256 authentication token for API requests
func (c *Client) GenerateAuthToken(requestBody string) (string, error) {
	// Step 1: Create metadata
	metadata := map[string]string{
		"id": c.appID,
	}
	
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	// Step 2: Encode metadata to Base64 and clean it
	metadataEncoded := base64.StdEncoding.EncodeToString(metadataJSON)
	metadataEncoded = cleanBase64String(metadataEncoded)
	
	// Step 3: Generate first signature (appKey + devKey)
	h := hmac.New(sha256.New, []byte(c.devKey))
	h.Write([]byte(c.appKey))
	firstSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// Step 4: Generate second signature (requestBody + firstSignature)
	h = hmac.New(sha256.New, []byte(firstSignature))
	h.Write([]byte(requestBody))
	secondSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	secondSignature = cleanBase64String(secondSignature)
	
	// Step 5: Combine metadata and signature
	token := metadataEncoded + "." + secondSignature
	
	logrus.Debugf("Generated OtomaX token for request body length: %d", len(requestBody))
	
	return token, nil
}

// cleanBase64String cleans Base64 string according to OtomaX requirements
func cleanBase64String(s string) string {
	// Remove padding (=), replace / with _, replace + with -
	s = string(bytes.ReplaceAll([]byte(s), []byte("="), []byte("")))
	s = string(bytes.ReplaceAll([]byte(s), []byte("/"), []byte("_")))
	s = string(bytes.ReplaceAll([]byte(s), []byte("+"), []byte("-")))
	return s
}

// makeRequest makes HTTP request to OtomaX API with authentication
func (c *Client) makeRequest(ctx context.Context, endpoint string, requestBody interface{}) (*http.Response, error) {
	// Marshal request body
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	// Generate authentication token
	token, err := c.GenerateAuthToken(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}
	
	// Create HTTP request
	url := c.baseURL + endpoint + "?token=" + token
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WhatsApp-Center/1.0")
	
	// Debug: Log request details (simplified)
	logrus.Debugf("OtomaX API Request - URL: %s", url)
	logrus.Debugf("OtomaX API Request - Body: %s", string(bodyBytes))
	
	// Make request
	logrus.Debugf("Making OtomaX API request to: %s", endpoint)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	
	return resp, nil
}

// parseResponse parses HTTP response and returns structured response
func (c *Client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Debug: Log raw response from OtomaX API
	logrus.Infof("=== OTOMAX API RESPONSE ===")
	logrus.Infof("Status: %d %s", resp.StatusCode, resp.Status)
	logrus.Infof("Raw Response Body: %s", string(body))
	logrus.Infof("=== END OTOMAX API RESPONSE ===")
	
	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logrus.Errorf("OtomaX API error: %s, status: %d, body: %s", resp.Status, resp.StatusCode, string(body))
		return pkgError.ExternalAPIError(fmt.Sprintf("OtomaX API error: %s", resp.Status))
	}
	
	// Parse JSON response
	if err := json.Unmarshal(body, result); err != nil {
		logrus.Errorf("Failed to parse OtomaX API response: %v, body: %s", err, string(body))
		return fmt.Errorf("failed to parse API response: %w", err)
	}
	
	logrus.Debugf("OtomaX API response parsed successfully")
	return nil
}

// InsertInbox sends message to OtomaX inbox
func (c *Client) InsertInbox(ctx context.Context, request domainOtomax.InsertInboxRequest) (*domainOtomax.InsertInboxResponse, error) {
	var response domainOtomax.InsertInboxResponse
	
	resp, err := c.makeRequest(ctx, "InsertInbox", request)
	if err != nil {
		return nil, err
	}
	
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}
	
	// Log parsed response details
	logrus.Infof("OtomaX InsertInbox Success - Ok: %t, KodeInbox: %d, Status: %d, StatusDesc: %s", 
		response.Ok, response.Result.KodeInbox, response.Result.Status, response.Result.StatusDesc)
	
	return &response, nil
}

// SetOutboxCallback sets the callback URL for outbox responses
func (c *Client) SetOutboxCallback(ctx context.Context, request domainOtomax.SetOutboxCallbackRequest) (*domainOtomax.SetOutboxCallbackResponse, error) {
	var response domainOtomax.SetOutboxCallbackResponse
	
	resp, err := c.makeRequest(ctx, "SetOutboxCallback", request)
	if err != nil {
		return nil, err
	}
	
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}
	
	logrus.Infof("Successfully set OtomaX callback URL: %s", request.URL)
	return &response, nil
}

// GetOutboxCallback gets the current callback URL
func (c *Client) GetOutboxCallback(ctx context.Context) (*domainOtomax.GetOutboxCallbackResponse, error) {
	var response domainOtomax.GetOutboxCallbackResponse
	
	resp, err := c.makeRequest(ctx, "GetOutboxCallback", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}
	
	return &response, nil
}

// GetRs gets reseller information
func (c *Client) GetRs(ctx context.Context, request domainOtomax.GetRsRequest) (*domainOtomax.GetRsResponse, error) {
	var response domainOtomax.GetRsResponse
	
	resp, err := c.makeRequest(ctx, "GetRs", request)
	if err != nil {
		return nil, err
	}
	
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}
	
	return &response, nil
}

// GetSaldoRs gets reseller balance
func (c *Client) GetSaldoRs(ctx context.Context, request domainOtomax.GetSaldoRsRequest) (*domainOtomax.GetSaldoRsResponse, error) {
	var response domainOtomax.GetSaldoRsResponse
	
	resp, err := c.makeRequest(ctx, "GetSaldoRs", request)
	if err != nil {
		return nil, err
	}
	
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}
	
	return &response, nil
}

// Test tests the API connection
func (c *Client) Test(ctx context.Context, request domainOtomax.TestRequest) (*domainOtomax.TestResponse, error) {
	var response domainOtomax.TestResponse
	
	resp, err := c.makeRequest(ctx, "Test", request)
	if err != nil {
		return nil, err
	}
	
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}
	
	return &response, nil
}
