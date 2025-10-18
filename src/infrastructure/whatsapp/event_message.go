package whatsapp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainOtomax "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/otomax"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/otomax"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// forwardMessageToWebhook is a helper function to forward message event to webhook url
func forwardMessageToWebhook(ctx context.Context, evt *events.Message) error {
	payload, err := createMessagePayload(ctx, evt)
	if err != nil {
		return err
	}

	return forwardPayloadToConfiguredWebhooks(ctx, payload, "message event")
}

// createOtomaxInsertInboxRequest creates request for OtomaX InsertInbox API
func createOtomaxInsertInboxRequest(ctx context.Context, evt *events.Message) (domainOtomax.InsertInboxRequest, error) {
	// Extract message text
	messageText := utils.ExtractMessageTextFromProto(evt.Message)
	
	// Extract phone number from JID (remove @s.whatsapp.net)
	senderJID := evt.Info.Sender.String()
	if senderJID == "" {
		return domainOtomax.InsertInboxRequest{}, fmt.Errorf("failed to get sender JID from message event")
	}
	
	// Remove @s.whatsapp.net suffix and :18 suffix to get just the phone number
	phoneNumber := strings.Split(senderJID, "@")[0]
	phoneNumber = strings.Split(phoneNumber, ":")[0] // Remove :18 suffix
	
	// Create OtomaX InsertInbox request
	request := domainOtomax.InsertInboxRequest{
		Pesan:        messageText,
		Pengirim:     phoneNumber, // Use phone number only: 6281295749258
		TipePengirim: "W",         // W for WhatsApp
		KodeTerminal: config.OtomaxDefaultKodeTerminal,
	}
	
	logrus.Debugf("Created OtomaX InsertInbox request: %+v", request)
	return request, nil
}

// forwardMessageToOtomax forwards WhatsApp message to OtomaX via InsertInbox
func forwardMessageToOtomax(ctx context.Context, evt *events.Message) error {
	// Check if OtomaX integration is enabled
	if !config.OtomaxEnabled {
		logrus.Debugf("OtomaX integration is disabled, skipping message forwarding")
		return nil
	}
	
	// Skip if message is from us (outgoing messages)
	if evt.Info.IsFromMe {
		return nil
	}
	
	// Skip group messages if not configured to forward groups
	if !config.OtomaxForwardGroups && utils.IsGroupJID(evt.Info.Chat.String()) {
		logrus.Debugf("Skipping group message for OtomaX forwarding")
		return nil
	}
	
	// Skip messages without text content
	messageText := utils.ExtractMessageTextFromProto(evt.Message)
	if strings.TrimSpace(messageText) == "" {
		logrus.Debugf("Skipping message without text content for OtomaX forwarding")
		return nil
	}
	
	// Create OtomaX InsertInbox request
	request, err := createOtomaxInsertInboxRequest(ctx, evt)
	if err != nil {
		logrus.Errorf("Failed to create OtomaX request: %v", err)
		return err
	}
	
	// Forward to OtomaX InsertInbox endpoint
	logrus.Infof("Forwarding WhatsApp message to OtomaX InsertInbox: sender=%s, message=%s", 
		evt.Info.Sender.String(), messageText)
	
	// Use OtomaX client to send request
	err = sendRequestToOtomaxClient(ctx, request, evt.Info.Sender.String())
	if err != nil {
		logrus.Errorf("Failed to send request to OtomaX: %v", err)
		return err
	}
	
	logrus.Infof("Successfully sent message to OtomaX InsertInbox")
	
	return nil
}

// sendRequestToOtomaxClient sends request to OtomaX using the existing client
func sendRequestToOtomaxClient(ctx context.Context, request domainOtomax.InsertInboxRequest, originalSenderJID string) error {
	// Create OtomaX client
	client := otomax.NewOtomaxClient()
	
	// Send request to OtomaX InsertInbox
	response, err := client.InsertInbox(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to send request to OtomaX: %w", err)
	}
	
	logrus.Infof("OtomaX InsertInbox response: %+v", response)
	
	// Check if status requires auto reply (21: Success with reason, 41: Bukan Reseller, 42: Format Salah)
	if response.Result.Status == 21 || response.Result.Status == 41 || response.Result.Status == 42 {
		var replyMessage string
		
		if response.Result.Status == 21 {
			// For status 21, use the "pesan" field if available, otherwise use StatusDesc
			if response.Result.Pesan != "" {
				replyMessage = response.Result.Pesan
			} else {
				replyMessage = response.Result.StatusDesc
			}
			logrus.Infof("Status %d requires auto reply with pesan: %s", response.Result.Status, replyMessage)
		} else {
			// For status 41 and 42, use StatusDesc
			replyMessage = response.Result.StatusDesc
			logrus.Infof("Status %d requires auto reply: %s", response.Result.Status, replyMessage)
		}
		
	// Send auto reply message to WhatsApp using original sender JID
	err = sendAutoReplyToWhatsAppWithJID(ctx, originalSenderJID, replyMessage)
		if err != nil {
			logrus.Errorf("Failed to send auto reply to WhatsApp: %v", err)
			// Don't return error here, just log it
		} else {
			logrus.Infof("Auto reply sent successfully to %s: %s", request.Pengirim, replyMessage)
		}
	}
	
	return nil
}

// sendAutoReplyToWhatsAppWithJID sends auto reply message using specific JID
func sendAutoReplyToWhatsAppWithJID(ctx context.Context, senderJID, statusDesc string) error {
	// Remove device part from JID (e.g., :18) to get clean user JID
	cleanJID := strings.Split(senderJID, ":")[0] // Remove :18 part
	if !strings.HasSuffix(cleanJID, "@s.whatsapp.net") {
		cleanJID = cleanJID + "@s.whatsapp.net"
	}
	
	// Parse the clean JID
	recipientJID, err := types.ParseJID(cleanJID)
	if err != nil {
		logrus.Errorf("Failed to parse clean JID %s: %v", cleanJID, err)
		return fmt.Errorf("failed to parse clean JID: %w", err)
	}
	
	// Send the auto-reply message using direct WhatsApp client
	response, err := cli.SendMessage(
		ctx,
		recipientJID,
		&waE2E.Message{Conversation: proto.String(statusDesc)},
	)
	
	if err != nil {
		logrus.Errorf("Failed to send OtomaX auto-reply message: %v", err)
		return err
	}
	
	logrus.Infof("OtomaX auto reply sent successfully to %s (from %s): %s (Message ID: %s)", cleanJID, senderJID, statusDesc, response.ID)
	return nil
}

// sendAutoReplyToWhatsApp sends auto reply message to WhatsApp user using direct WhatsApp client
func sendAutoReplyToWhatsApp(ctx context.Context, phoneNumber, statusDesc string) error {
	// Format recipient JID
	recipientJID := utils.FormatJID(phoneNumber + "@s.whatsapp.net")
	
	// Send the auto-reply message using direct WhatsApp client
	response, err := cli.SendMessage(
		ctx,
		recipientJID,
		&waE2E.Message{Conversation: proto.String(statusDesc)},
	)
	
	if err != nil {
		logrus.Errorf("Failed to send OtomaX auto-reply message: %v", err)
		return err
	}
	
	logrus.Infof("OtomaX auto reply sent successfully to %s: %s (Message ID: %s)", phoneNumber, statusDesc, response.ID)
	
	// TODO: Add message storage logic here if needed
	// Similar to existing auto reply storage logic
	
	return nil
}

func createMessagePayload(ctx context.Context, evt *events.Message) (map[string]any, error) {
	message := utils.BuildEventMessage(evt)
	waReaction := utils.BuildEventReaction(evt)
	forwarded := utils.BuildForwarded(evt)

	body := make(map[string]any)

	body["sender_id"] = evt.Info.Sender.User
	body["chat_id"] = evt.Info.Chat.User

	if from := evt.Info.SourceString(); from != "" {
		body["from"] = from

		from_user, from_group := from, ""
		if strings.Contains(from, " in ") {
			from_user = strings.Split(from, " in ")[0]
			from_group = strings.Split(from, " in ")[1]
		}

		if strings.HasSuffix(from_user, "@lid") {
			body["from_lid"] = from_user
			lid, err := types.ParseJID(from_user)
			if err != nil {
				logrus.Errorf("Error when parse jid: %v", err)
			} else {
				pn, err := cli.Store.LIDs.GetPNForLID(ctx, lid)
				if err != nil {
					logrus.Errorf("Error when get pn for lid %s: %v", lid.String(), err)
				}
				if !pn.IsEmpty() {
					if from_group != "" {
						body["from"] = fmt.Sprintf("%s in %s", pn.String(), from_group)
					} else {
						body["from"] = pn.String()
					}
				}
			}
		}
	}
	if message.ID != "" {
		tags := regexp.MustCompile(`\B@\w+`).FindAllString(message.Text, -1)
		tagsMap := make(map[string]bool)
		for _, tag := range tags {
			tagsMap[tag] = true
		}
		for tag := range tagsMap {
			lid, err := types.ParseJID(tag[1:] + "@lid")
			if err != nil {
				logrus.Errorf("Error when parse jid: %v", err)
			} else {
				pn, err := cli.Store.LIDs.GetPNForLID(ctx, lid)
				if err != nil {
					logrus.Errorf("Error when get pn for lid %s: %v", lid.String(), err)
				}
				if !pn.IsEmpty() {
					message.Text = strings.Replace(message.Text, tag, fmt.Sprintf("@%s", pn.User), -1)
				}
			}
		}
		body["message"] = message
	}
	if pushname := evt.Info.PushName; pushname != "" {
		body["pushname"] = pushname
	}
	if waReaction.Message != "" {
		body["reaction"] = waReaction
	}
	if evt.IsViewOnce {
		body["view_once"] = evt.IsViewOnce
	}
	if forwarded {
		body["forwarded"] = forwarded
	}
	if timestamp := evt.Info.Timestamp.Format(time.RFC3339); timestamp != "" {
		body["timestamp"] = timestamp
	}

	// Handle protocol messages (revoke, etc.)
	if protocolMessage := evt.Message.GetProtocolMessage(); protocolMessage != nil {
		protocolType := protocolMessage.GetType().String()

		switch protocolType {
		case "REVOKE":
			body["action"] = "message_revoked"
			if key := protocolMessage.GetKey(); key != nil {
				body["revoked_message_id"] = key.GetID()
				body["revoked_from_me"] = key.GetFromMe()
				if key.GetRemoteJID() != "" {
					body["revoked_chat"] = key.GetRemoteJID()
				}
			}
		case "MESSAGE_EDIT":
			body["action"] = "message_edited"
			if editedMessage := protocolMessage.GetEditedMessage(); editedMessage != nil {
				if editedText := editedMessage.GetExtendedTextMessage(); editedText != nil {
					body["edited_text"] = editedText.GetText()
				} else if editedConv := editedMessage.GetConversation(); editedConv != "" {
					body["edited_text"] = editedConv
				}
			}
		}
	}

	if audioMedia := evt.Message.GetAudioMessage(); audioMedia != nil {
		path, err := utils.ExtractMedia(ctx, cli, config.PathMedia, audioMedia)
		if err != nil {
			logrus.Errorf("Failed to download audio from %s: %v", evt.Info.SourceString(), err)
			return nil, pkgError.WebhookError(fmt.Sprintf("Failed to download audio: %v", err))
		}
		body["audio"] = path
	}

	if contactMessage := evt.Message.GetContactMessage(); contactMessage != nil {
		body["contact"] = contactMessage
	}

	if documentMedia := evt.Message.GetDocumentMessage(); documentMedia != nil {
		path, err := utils.ExtractMedia(ctx, cli, config.PathMedia, documentMedia)
		if err != nil {
			logrus.Errorf("Failed to download document from %s: %v", evt.Info.SourceString(), err)
			return nil, pkgError.WebhookError(fmt.Sprintf("Failed to download document: %v", err))
		}
		body["document"] = path
	}

	if imageMedia := evt.Message.GetImageMessage(); imageMedia != nil {
		path, err := utils.ExtractMedia(ctx, cli, config.PathMedia, imageMedia)
		if err != nil {
			logrus.Errorf("Failed to download image from %s: %v", evt.Info.SourceString(), err)
			return nil, pkgError.WebhookError(fmt.Sprintf("Failed to download image: %v", err))
		}
		body["image"] = path
	}

	if listMessage := evt.Message.GetListMessage(); listMessage != nil {
		body["list"] = listMessage
	}

	if liveLocationMessage := evt.Message.GetLiveLocationMessage(); liveLocationMessage != nil {
		body["live_location"] = liveLocationMessage
	}

	if locationMessage := evt.Message.GetLocationMessage(); locationMessage != nil {
		body["location"] = locationMessage
	}

	if orderMessage := evt.Message.GetOrderMessage(); orderMessage != nil {
		body["order"] = orderMessage
	}

	if stickerMedia := evt.Message.GetStickerMessage(); stickerMedia != nil {
		path, err := utils.ExtractMedia(ctx, cli, config.PathMedia, stickerMedia)
		if err != nil {
			logrus.Errorf("Failed to download sticker from %s: %v", evt.Info.SourceString(), err)
			return nil, pkgError.WebhookError(fmt.Sprintf("Failed to download sticker: %v", err))
		}
		body["sticker"] = path
	}

	if videoMedia := evt.Message.GetVideoMessage(); videoMedia != nil {
		path, err := utils.ExtractMedia(ctx, cli, config.PathMedia, videoMedia)
		if err != nil {
			logrus.Errorf("Failed to download video from %s: %v", evt.Info.SourceString(), err)
			return nil, pkgError.WebhookError(fmt.Sprintf("Failed to download video: %v", err))
		}
		body["video"] = path
	}

	return body, nil
}
