package cmd

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"os"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	domainChat "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chat"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	domainGroup "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/group"
	domainMessage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/message"
	domainNewsletter "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/newsletter"
	domainOtomax "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/otomax"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	domainUser "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/user"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/otomax"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mau.fi/whatsmeow"
)

var (
	EmbedIndex embed.FS
	EmbedViews embed.FS

	// Whatsapp
	whatsappCli *whatsmeow.Client

	// Chat Storage
	chatStorageDB   *sql.DB
	chatStorageRepo domainChatStorage.IChatStorageRepository

	// Usecase
	appUsecase        domainApp.IAppUsecase
	chatUsecase       domainChat.IChatUsecase
	sendUsecase       domainSend.ISendUsecase
	userUsecase       domainUser.IUserUsecase
	messageUsecase    domainMessage.IMessageUsecase
	groupUsecase      domainGroup.IGroupUsecase
	newsletterUsecase domainNewsletter.INewsletterUsecase
	otomaxUsecase     domainOtomax.IOtomaxUsecase
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "Send free whatsapp API",
	Long: `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice, 
you can send whatsapp over http api but your whatsapp account have to be multi device version`,
}

func init() {
	// Load environment variables first
	utils.LoadConfig(".")

	time.Local = time.UTC

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Initialize flags first, before any subcommands are added
	initFlags()

	// Then initialize other components
	cobra.OnInitialize(initEnvConfig, initApp)
}

// initEnvConfig loads configuration from environment variables
func initEnvConfig() {
	fmt.Println(viper.AllSettings())
	// Application settings
	if envPort := viper.GetString("app_port"); envPort != "" {
		config.AppPort = envPort
	}
	if envDebug := viper.GetBool("app_debug"); envDebug {
		config.AppDebug = envDebug
	}
	if envOs := viper.GetString("app_os"); envOs != "" {
		config.AppOs = envOs
	}
	if envBasicAuth := viper.GetString("app_basic_auth"); envBasicAuth != "" {
		credential := strings.Split(envBasicAuth, ",")
		config.AppBasicAuthCredential = credential
	}
	if envBasePath := viper.GetString("app_base_path"); envBasePath != "" {
		config.AppBasePath = envBasePath
	}

	// Database settings
	if envDBURI := viper.GetString("db_uri"); envDBURI != "" {
		config.DBURI = envDBURI
	}
	if envDBKEYSURI := viper.GetString("db_keys_uri"); envDBKEYSURI != "" {
		config.DBKeysURI = envDBKEYSURI
	}

	// WhatsApp settings
	if envAutoReply := viper.GetString("whatsapp_auto_reply"); envAutoReply != "" {
		config.WhatsappAutoReplyMessage = envAutoReply
	}
	if viper.IsSet("whatsapp_auto_mark_read") {
		config.WhatsappAutoMarkRead = viper.GetBool("whatsapp_auto_mark_read")
	}
	if envWebhook := viper.GetString("whatsapp_webhook"); envWebhook != "" {
		webhook := strings.Split(envWebhook, ",")
		config.WhatsappWebhook = webhook
	}
	if envWebhookSecret := viper.GetString("whatsapp_webhook_secret"); envWebhookSecret != "" {
		config.WhatsappWebhookSecret = envWebhookSecret
	}
	if viper.IsSet("whatsapp_account_validation") {
		config.WhatsappAccountValidation = viper.GetBool("whatsapp_account_validation")
	}

	// OtomaX settings
	if viper.IsSet("otomax_enabled") {
		config.OtomaxEnabled = viper.GetBool("otomax_enabled")
	}
	if envOtomaxAPIURL := viper.GetString("otomax_api_url"); envOtomaxAPIURL != "" {
		config.OtomaxAPIURL = envOtomaxAPIURL
	}
	if envOtomaxAppID := viper.GetString("otomax_app_id"); envOtomaxAppID != "" {
		config.OtomaxAppID = envOtomaxAppID
	}
	if envOtomaxAppKey := viper.GetString("otomax_app_key"); envOtomaxAppKey != "" {
		config.OtomaxAppKey = envOtomaxAppKey
	}
	if envOtomaxDevKey := viper.GetString("otomax_dev_key"); envOtomaxDevKey != "" {
		config.OtomaxDevKey = envOtomaxDevKey
	}
	if envOtomaxDefaultReseller := viper.GetString("otomax_default_reseller"); envOtomaxDefaultReseller != "" {
		config.OtomaxDefaultReseller = envOtomaxDefaultReseller
	}
	if viper.IsSet("otomax_forward_incoming") {
		config.OtomaxForwardIncoming = viper.GetBool("otomax_forward_incoming")
	}
	if viper.IsSet("otomax_forward_outgoing") {
		config.OtomaxForwardOutgoing = viper.GetBool("otomax_forward_outgoing")
	}
	if viper.IsSet("otomax_forward_groups") {
		config.OtomaxForwardGroups = viper.GetBool("otomax_forward_groups")
	}
	if viper.IsSet("otomax_forward_media") {
		config.OtomaxForwardMedia = viper.GetBool("otomax_forward_media")
	}
	if viper.IsSet("otomax_auto_reply_enabled") {
		config.OtomaxAutoReplyEnabled = viper.GetBool("otomax_auto_reply_enabled")
	}
}

func initFlags() {
	// Application flags
	rootCmd.PersistentFlags().StringVarP(
		&config.AppPort,
		"port", "p",
		config.AppPort,
		"change port number with --port <number> | example: --port=8080",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&config.AppDebug,
		"debug", "d",
		config.AppDebug,
		"hide or displaying log with --debug <true/false> | example: --debug=true",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppOs,
		"os", "",
		config.AppOs,
		`os name --os <string> | example: --os="Chrome"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppBasicAuthCredential,
		"basic-auth", "b",
		config.AppBasicAuthCredential,
		"basic auth credential | -b=yourUsername:yourPassword",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppBasePath,
		"base-path", "",
		config.AppBasePath,
		`base path for subpath deployment --base-path <string> | example: --base-path="/gowa"`,
	)

	// Database flags
	rootCmd.PersistentFlags().StringVarP(
		&config.DBURI,
		"db-uri", "",
		config.DBURI,
		`the database uri to store the connection data database uri (by default, we'll use sqlite3 under storages/whatsapp.db). database uri --db-uri <string> | example: --db-uri="file:storages/whatsapp.db?_foreign_keys=on or postgres://user:password@localhost:5432/whatsapp"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.DBKeysURI,
		"db-keys-uri", "",
		config.DBKeysURI,
		`the database uri to store the keys database uri (by default, we'll use the same database uri). database uri --db-keys-uri <string> | example: --db-keys-uri="file::memory:?cache=shared&_foreign_keys=on"`,
	)

	// WhatsApp flags
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappAutoReplyMessage,
		"autoreply", "",
		config.WhatsappAutoReplyMessage,
		`auto reply when received message --autoreply <string> | example: --autoreply="Don't reply this message"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAutoMarkRead,
		"auto-mark-read", "",
		config.WhatsappAutoMarkRead,
		`auto mark incoming messages as read --auto-mark-read <true/false> | example: --auto-mark-read=true`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.WhatsappWebhook,
		"webhook", "w",
		config.WhatsappWebhook,
		`forward event to webhook --webhook <string> | example: --webhook="https://yourcallback.com/callback"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappWebhookSecret,
		"webhook-secret", "",
		config.WhatsappWebhookSecret,
		`secure webhook request --webhook-secret <string> | example: --webhook-secret="super-secret-key"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAccountValidation,
		"account-validation", "",
		config.WhatsappAccountValidation,
		`enable or disable account validation --account-validation <true/false> | example: --account-validation=true`,
	)

	// OtomaX flags
	rootCmd.PersistentFlags().BoolVarP(
		&config.OtomaxEnabled,
		"otomax-enabled", "",
		config.OtomaxEnabled,
		`enable or disable OtomaX integration --otomax-enabled <true/false> | example: --otomax-enabled=true`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.OtomaxAPIURL,
		"otomax-api-url", "",
		config.OtomaxAPIURL,
		`OtomaX API URL --otomax-api-url <string> | example: --otomax-api-url="http://localhost:5000/"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.OtomaxAppID,
		"otomax-app-id", "",
		config.OtomaxAppID,
		`OtomaX App ID --otomax-app-id <string> | example: --otomax-app-id="OtomaX.Addon"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.OtomaxAppKey,
		"otomax-app-key", "",
		config.OtomaxAppKey,
		`OtomaX App Key --otomax-app-key <string> | example: --otomax-app-key="demoKey"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.OtomaxDevKey,
		"otomax-dev-key", "",
		config.OtomaxDevKey,
		`OtomaX Dev Key --otomax-dev-key <string> | example: --otomax-dev-key="YAR.OtomaX.OpenApi.App"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.OtomaxDefaultReseller,
		"otomax-default-reseller", "",
		config.OtomaxDefaultReseller,
		`OtomaX Default Reseller --otomax-default-reseller <string> | example: --otomax-default-reseller="yusuf"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.OtomaxForwardIncoming,
		"otomax-forward-incoming", "",
		config.OtomaxForwardIncoming,
		`forward incoming messages to OtomaX --otomax-forward-incoming <true/false> | example: --otomax-forward-incoming=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.OtomaxForwardOutgoing,
		"otomax-forward-outgoing", "",
		config.OtomaxForwardOutgoing,
		`forward outgoing messages to OtomaX --otomax-forward-outgoing <true/false> | example: --otomax-forward-outgoing=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.OtomaxForwardGroups,
		"otomax-forward-groups", "",
		config.OtomaxForwardGroups,
		`forward group messages to OtomaX --otomax-forward-groups <true/false> | example: --otomax-forward-groups=false`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.OtomaxForwardMedia,
		"otomax-forward-media", "",
		config.OtomaxForwardMedia,
		`forward media messages to OtomaX --otomax-forward-media <true/false> | example: --otomax-forward-media=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.OtomaxAutoReplyEnabled,
		"otomax-auto-reply-enabled", "",
		config.OtomaxAutoReplyEnabled,
		`enable auto reply for OtomaX --otomax-auto-reply-enabled <true/false> | example: --otomax-auto-reply-enabled=true`,
	)
}

func initChatStorage() (*sql.DB, error) {
	connStr := fmt.Sprintf("%s?_journal_mode=WAL", config.ChatStorageURI)
	if config.ChatStorageEnableForeignKeys {
		connStr += "&_foreign_keys=on"
	}

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initApp() {
	if config.AppDebug {
		config.WhatsappLogLevel = "DEBUG"
		logrus.SetLevel(logrus.DebugLevel)
	}

	//preparing folder if not exist
	err := utils.CreateFolder(config.PathQrCode, config.PathSendItems, config.PathStorages, config.PathMedia)
	if err != nil {
		logrus.Errorln(err)
	}

	ctx := context.Background()

	chatStorageDB, err = initChatStorage()
	if err != nil {
		// Terminate the application if chat storage fails to initialize to avoid nil pointer panics later.
		logrus.Fatalf("failed to initialize chat storage: %v", err)
	}

	chatStorageRepo = chatstorage.NewStorageRepository(chatStorageDB)
	chatStorageRepo.InitializeSchema()

	whatsappDB := whatsapp.InitWaDB(ctx, config.DBURI)
	var keysDB *sqlstore.Container
	if config.DBKeysURI != "" {
		keysDB = whatsapp.InitWaDB(ctx, config.DBKeysURI)
	}

	whatsapp.InitWaCLI(ctx, whatsappDB, keysDB, chatStorageRepo)

	// Usecase
	appUsecase = usecase.NewAppService(chatStorageRepo)
	chatUsecase = usecase.NewChatService(chatStorageRepo)
	sendUsecase = usecase.NewSendService(appUsecase, chatStorageRepo)
	userUsecase = usecase.NewUserService()
	messageUsecase = usecase.NewMessageService(chatStorageRepo)
	groupUsecase = usecase.NewGroupService()
	newsletterUsecase = usecase.NewNewsletterService()

	// Initialize OtomaX service if enabled
	if config.OtomaxEnabled {
		otomaxClient := otomax.NewOtomaxClient()
		otomaxUsecase = usecase.NewOtomaxService(otomaxClient, sendUsecase)
		
		// Set OtomaX service in WhatsApp infrastructure for message processing
		whatsapp.SetOtomaxService(otomaxUsecase)
		
		logrus.Infof("OtomaX integration initialized with API URL: %s", config.OtomaxAPIURL)
	} else {
		logrus.Info("OtomaX integration is disabled")
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(embedIndex embed.FS, embedViews embed.FS) {
	EmbedIndex = embedIndex
	EmbedViews = embedViews
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
