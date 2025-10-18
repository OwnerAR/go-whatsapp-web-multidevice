package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	domainOtomax "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/otomax"
	"github.com/gofiber/fiber/v2"
)

type Otomax struct {
	Service domainOtomax.IOtomaxUsecase
}

// InitRestOtomax initializes OtomaX REST endpoints
func InitRestOtomax(app fiber.Router, service domainOtomax.IOtomaxUsecase) Otomax {
	rest := Otomax{Service: service}
	
	// OtomaX API endpoints
	app.Post("/otomax/insert-inbox", rest.InsertInbox)
	app.Post("/otomax/set-callback", rest.SetCallbackURL)
	app.Get("/otomax/get-callback", rest.GetCallbackURL)
	app.Post("/otomax/test", rest.TestConnection)
	app.Get("/otomax/reseller/:kode", rest.GetResellerInfo)
	app.Get("/otomax/reseller/:kode/balance", rest.GetResellerBalance)
	app.Get("/otomax/health", rest.HealthCheck)
	
	// Callback endpoint for OtomaX responses
	app.Post("/otomax/callback", rest.HandleCallback)
	
	return rest
}

// InsertInbox handles InsertInbox requests
func (controller *Otomax) InsertInbox(c *fiber.Ctx) error {
	var request domainOtomax.InsertInboxRequest
	err := c.BodyParser(&request)
	utils.PanicIfNeeded(err)

	response, err := controller.Service.SendMessageToOtomax(c.UserContext(), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Message sent to OtomaX successfully",
		Results: response,
	})
}

// SetCallbackURL handles SetOutboxCallback requests
func (controller *Otomax) SetCallbackURL(c *fiber.Ctx) error {
	var request domainOtomax.SetOutboxCallbackRequest
	err := c.BodyParser(&request)
	utils.PanicIfNeeded(err)

	response, err := controller.Service.SetCallbackURL(c.UserContext(), request)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Callback URL set successfully",
		Results: response,
	})
}

// GetCallbackURL handles GetOutboxCallback requests
func (controller *Otomax) GetCallbackURL(c *fiber.Ctx) error {
	response, err := controller.Service.GetCallbackURL(c.UserContext())
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Callback URL retrieved successfully",
		Results: response,
	})
}

// TestConnection handles Test requests
func (controller *Otomax) TestConnection(c *fiber.Ctx) error {
	var request domainOtomax.TestRequest
	err := c.BodyParser(&request)
	utils.PanicIfNeeded(err)

	response, err := controller.Service.TestConnection(c.UserContext(), request.Phone)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Connection test completed",
		Results: response,
	})
}

// GetResellerInfo handles GetRs requests
func (controller *Otomax) GetResellerInfo(c *fiber.Ctx) error {
	resellerCode := c.Params("kode")
	if resellerCode == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Reseller code is required",
		})
	}

	response, err := controller.Service.GetResellerInfo(c.UserContext(), resellerCode)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Reseller information retrieved successfully",
		Results: response,
	})
}

// GetResellerBalance handles GetSaldoRs requests
func (controller *Otomax) GetResellerBalance(c *fiber.Ctx) error {
	resellerCode := c.Params("kode")
	if resellerCode == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Reseller code is required",
		})
	}

	response, err := controller.Service.GetResellerBalance(c.UserContext(), resellerCode)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Reseller balance retrieved successfully",
		Results: response,
	})
}

// HealthCheck provides health status of OtomaX integration
func (controller *Otomax) HealthCheck(c *fiber.Ctx) error {
	// Test connection with a dummy phone number
	response, err := controller.Service.TestConnection(c.UserContext(), "08123456789")
	
	var status string
	var message string
	var data interface{}
	
	if err != nil {
		status = "ERROR"
		message = "OtomaX integration is not healthy"
		data = map[string]interface{}{
			"error": err.Error(),
		}
	} else {
		status = "HEALTHY"
		message = "OtomaX integration is healthy"
		data = response
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    status,
		Message: message,
		Results: data,
	})
}

// HandleCallback handles callback responses from OtomaX
func (controller *Otomax) HandleCallback(c *fiber.Ctx) error {
	var payload domainOtomax.CallbackPayload
	err := c.BodyParser(&payload)
	utils.PanicIfNeeded(err)

	err = controller.Service.HandleOtomaxCallback(c.UserContext(), payload)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Callback processed successfully",
		Results: map[string]interface{}{
			"kode":   payload.Kode,
			"status": payload.Status,
		},
	})
}
