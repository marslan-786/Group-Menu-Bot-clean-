package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	// "go.mau.fi/whatsmeow/types" // (Uncomment if needed later)
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// 🛑 Basic Configuration
const (
	BOT_NAME   = "Impossible Bot"
	OWNER_NAME = "Nothing Is Impossible"
)

// 🛡️ Anti-Spam Variables
var RestrictedGroups = make(map[string]bool)
var AuthorizedBots = make(map[string]bool)

// =========================================================================
// ⚡ MAIN PROCESSOR (Lightweight & Fast)
// =========================================================================
func processMessage(client *whatsmeow.Client, v *events.Message) {
	// 1. Panic Recovery
	defer recovery()

	// 2. Timestamp Check (5 seconds max delay allowed)
	if time.Since(v.Info.Timestamp) > 5*time.Second {
		return
	}

	// 3. Text Extraction
	bodyRaw := getText(v.Message)
	if bodyRaw == "" {
		return // Ignore empty messages
	}
	bodyClean := strings.TrimSpace(bodyRaw)

	// 4. Fast Bot ID
	rawBotID := client.Store.ID.User
	botID := getCleanID(rawBotID)

	// 5. Variables
	chatID := v.Info.Chat.String()
	// isGroup := v.Info.IsGroup // (Future use)

	// 6. Spam Filter
	if RestrictedGroups[chatID] && !AuthorizedBots[botID] {
		return
	}

	// 7. Get Prefix (Default is '.')
	prefix := getPrefix(botID)
	if !strings.HasPrefix(bodyClean, prefix) {
		return // Not a command
	}

	// 8. Command Parsing
	msgWithoutPrefix := strings.TrimPrefix(bodyClean, prefix)
	words := strings.Fields(msgWithoutPrefix)
	if len(words) == 0 { return }

	cmd := strings.ToLower(words[0])
	fullArgs := strings.TrimSpace(strings.Join(words[1:], " ")) // Arguments for setprefix

	fmt.Printf("🚀 [EXEC] Bot:%s | CMD:%s\n", botID, cmd)

	// =====================================================================
	// 🔥 COMMAND SWITCH (Background Execution)
	// =====================================================================
	go func() {
		defer recovery()

		switch cmd {
		// ✅ MENU COMMAND
		case "menu", "help", "list":
			react(client, v.Info.Chat, v.Info.ID, "📜")
			sendMenu(client, v, botID, prefix)
		
		// ✅ SET PREFIX COMMAND
		case "setprefix", "prefix":
			// صرف اونر استعمال کر سکے
			if !isOwner(client, v.Info.Sender) {
				replyMessage(client, v, "❌ Only Owner Command!")
				return
			}
			if fullArgs == "" {
				replyMessage(client, v, fmt.Sprintf("⚠️ Usage: %ssetprefix <symbol>\nExample: %ssetprefix !", prefix, prefix))
				return
			}
			
			// نیا پریفکس سیٹ کریں
			updatePrefixDB(botID, fullArgs)
			replyMessage(client, v, fmt.Sprintf("✅ Prefix updated to: [ %s ]", fullArgs))

		// 🛠️ باقی کمانڈز ہم بعد میں یہاں ایڈ کریں گے
		}
	}()
}

// =========================================================================
// 📜 MENU FUNCTION
// =========================================================================
func sendMenu(client *whatsmeow.Client, v *events.Message, botID, p string) {
	uptimeStr := getFormattedUptime()
	currentMode := "PUBLIC" 

	menu := fmt.Sprintf(`╔══════════════════════╗
║     ✨ %s ✨     
╠══════════════════════╣
║ 👋 *Assalam-o-Alaikum*
║ 👑 *Owner:* %s              
║ 🛡️ *Mode:* %s               
║ ⏳ *Uptime:* %s             
║ ⚡ *Prefix:* %s
╠══════════════════════╣
║ ╭──── SYSTEM ─────╮
║ │ 🔸 *%ssetprefix* - Change Symbol
║ │ 🔸 *%smenu* - Show this list
║ ╰──────────────────╯
╠══════════════════════╣
║ © 2025 Nothing is Impossible 
╚══════════════════════╝`,
		BOT_NAME, OWNER_NAME, currentMode, uptimeStr, p,
		p, p)

	// ✅ تصویر کے ساتھ بھیجیں
	imgData, err := os.ReadFile("pic.png")
	if err == nil {
		uploadResp, err := client.Upload(context.Background(), imgData, whatsmeow.MediaImage)
		if err == nil {
			imgMsg := &waProto.Message{
				ImageMessage: &waProto.ImageMessage{
					Caption:       proto.String(menu),
					URL:           proto.String(uploadResp.URL),
					DirectPath:    proto.String(uploadResp.DirectPath),
					MediaKey:      uploadResp.MediaKey,
					Mimetype:      proto.String("image/png"),
					FileEncSHA256: uploadResp.FileEncSHA256,
					FileSHA256:    uploadResp.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(imgData))),
				},
			}
			client.SendMessage(context.Background(), v.Info.Chat, imgMsg)
			return
		}
	}

	// اگر تصویر نہ ملے تو ٹیکسٹ بھیجیں
	replyMessage(client, v, menu)
}

// =========================================================================
// 🛠️ HELPER FUNCTIONS
// =========================================================================

func getText(msg *waProto.Message) string {
	if msg == nil { return "" }
	if msg.Conversation != nil { return *msg.Conversation }
	if msg.ExtendedTextMessage != nil { return *msg.ExtendedTextMessage.Text }
	if msg.ImageMessage != nil { return *msg.ImageMessage.Caption }
	if msg.VideoMessage != nil { return *msg.VideoMessage.Caption }
	return ""
}

func getCleanID(id string) string {
	if strings.Contains(id, ":") {
		id = strings.Split(id, ":")[0]
	}
	return strings.TrimSuffix(id, "@s.whatsapp.net")
}

// ✅ Default Prefix Logic
func getPrefix(botID string) string {
	prefixMutex.RLock()
	p, ok := botPrefixes[botID]
	prefixMutex.RUnlock()
	if ok && p != "" { return p }
	return "." // Default Prefix
}

// ✅ Update Prefix in Memory + Redis
func updatePrefixDB(botID string, newPrefix string) {
	// 1. Memory Update
	prefixMutex.Lock()
	botPrefixes[botID] = newPrefix
	prefixMutex.Unlock()

	// 2. Redis Update (if available)
	if rdb != nil {
		ctx := context.Background()
		rdb.Set(ctx, "prefix:"+botID, newPrefix, 0)
	}
}

func recovery() {
	if r := recover(); r != nil {
		fmt.Printf("⚠️ Panic Recovered: %v\n", r)
	}
}

func react(client *whatsmeow.Client, chatID types.JID, msgID types.MessageID, emoji string) {
	client.SendMessage(context.Background(), chatID, &waProto.Message{
		ReactionMessage: &waProto.ReactionMessage{
			Key: &waProto.MessageKey{
				RemoteJid: proto.String(chatID.String()),
				FromMe:    proto.Bool(false),
				Id:        proto.String(msgID),
			},
			Text: proto.String(emoji),
		},
	})
}

func replyMessage(client *whatsmeow.Client, v *events.Message, text string) {
	client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      proto.String(v.Info.ID),
				Participant:   proto.String(v.Info.Sender.String()),
				QuotedMessage: v.Message,
			},
		},
	})
}

func getFormattedUptime() string {
	// (Ensure persistentUptime is accessible from main package)
	duration := time.Duration(persistentUptime) * time.Second
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}

func isOwner(client *whatsmeow.Client, sender types.JID) bool {
	// Replace with your actual number
	return strings.Contains(sender.User, "923001234567") 
}