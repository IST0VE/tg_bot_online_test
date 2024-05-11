package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	lastKnownStatus = false
)

var lastCheck time.Time

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизация с аккаунта %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // Добавляем проверку на nil
			if update.Message.IsCommand() {
				handleCommands(bot, update)
			}
		}

		if update.MyChatMember != nil {
			handleMyChatMember(bot, update)
		}

		currentTime := time.Now()
		if currentTime.Sub(lastCheck) > time.Minute*1 {
			checkAndNotifyChannelPresence(bot)
			lastCheck = currentTime
		}
	}

}

func handleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	switch update.Message.Command() {
	case "status":
		msg.Text = fmt.Sprintf("Бот онлайн. Последняя проверка в %s", time.Now().Format(time.RFC1123))
	case "check":
		args := update.Message.CommandArguments()
		if args == "" {
			msg.Text = "Пожалуйста, укажите ID канала после команды, например: /check -1001234567890"
		} else {
			chatID, err := strconv.ParseInt(args, 10, 64)
			if err != nil {
				msg.Text = "Неправильный формат ID канала."
			} else {
				if isBotInChannel(bot, chatID) {
					msg.Text = "Бот находится в этом канале/группе."
				} else {
					msg.Text = "Бот не состоит в этом канале/группе или не может получить сообщения."
				}
			}
		}
	default:
		msg.Text = "Я не знаю эту команду."
	}

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func isBotInChannel(bot *tgbotapi.BotAPI, chatID int64) bool {
	chatInfoConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}

	_, err := bot.GetChat(chatInfoConfig)
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}

func checkAndNotifyChannelPresence(bot *tgbotapi.BotAPI) {
	chatID, err := strconv.ParseInt(os.Getenv("TARGET_CHAT_ID"), 10, 64)
	if err != nil {
		log.Printf("Ошибка при разборе TARGET_CHAT_ID: %v", err)
		return
	}

	isPresent := isBotInChannel(bot, chatID)
	if isPresent != lastKnownStatus {
		lastKnownStatus = isPresent
		notifyUser(bot, isPresent)
	}
}

func notifyUser(bot *tgbotapi.BotAPI, isPresent bool) {
	userID, err := strconv.ParseInt(os.Getenv("NOTIFICATION_USER_ID"), 10, 64)
	if err != nil {
		log.Printf("Ошибка при разборе NOTIFICATION_USER_ID: %v", err)
		return
	}

	msgText := ""
	if isPresent {
		msgText = "Бот добавлен в канал."
	} else {
		msgText = "Бот удален из канала."
	}

	msg := tgbotapi.NewMessage(userID, msgText)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке уведомления пользователю: %v", err)
	}
}

func handleMyChatMember(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.MyChatMember.NewChatMember.User.ID == bot.Self.ID {
		msgText := fmt.Sprintf("Статус бота изменён на: %s", update.MyChatMember.NewChatMember.Status)
		log.Println(msgText)

		if update.MyChatMember.NewChatMember.Status == "kicked" || update.MyChatMember.NewChatMember.Status == "left" {
			notifyUser(bot, false)
		}
	}
}
