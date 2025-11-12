package config

import (
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
)

var (
	AppVersion             = "v7.8.2"
	AppPort                = "3000"
	AppDebug               = false
	AppOs                  = "AldinoKemal"
	AppPlatform            = waCompanionReg.DeviceProps_PlatformType(1)
	AppBasicAuthCredential []string
	AppBasePath            = ""

	McpPort = "8080"
	McpHost = "localhost"

	PathQrCode    = "statics/qrcode"
	PathSendItems = "statics/senditems"
	PathMedia     = "statics/media"
	PathStorages  = "storages"

	DBURI     = "file:storages/whatsapp.db?_foreign_keys=on"
	DBKeysURI = ""

	WhatsappAutoReplyMessage       string
	WhatsappAutoMarkRead           = false // Auto-mark incoming messages as read
	WhatsappWebhook                []string
	WhatsappWebhookSecret                = "secret"
	WhatsappLogLevel                     = "ERROR"
	WhatsappSettingMaxImageSize    int64 = 20000000  // 20MB
	WhatsappSettingMaxFileSize     int64 = 50000000  // 50MB
	WhatsappSettingMaxVideoSize    int64 = 100000000 // 100MB
	WhatsappSettingMaxDownloadSize int64 = 500000000 // 500MB
	WhatsappTypeUser                     = "@s.whatsapp.net"
	WhatsappTypeGroup                    = "@g.us"
	WhatsappAccountValidation            = true

	ChatStorageURI               = "file:storages/chatstorage.db"
	ChatStorageEnableForeignKeys = true
	ChatStorageEnableWAL         = true

	// OtomaX API Configuration
	OtomaxEnabled               = false
	OtomaxAPIURL               = "http://localhost:5000/"
	OtomaxAppID                = "OtomaX.Addon"
	OtomaxAppKey               = "demoKey"
	OtomaxDevKey               = "YAR.OtomaX.OpenApi.App"
	OtomaxDefaultReseller      = ""
	OtomaxForwardIncoming      = true
	OtomaxForwardOutgoing      = true
	OtomaxForwardGroups        = false
	OtomaxForwardMedia         = true
	OtomaxAutoReplyEnabled     = true
	OtomaxDefaultKodeTerminal = 2
)
