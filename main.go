package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/joho/godotenv"
)

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
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			handleCommands(bot, update)
		}
	}
}

func handleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	switch update.Message.Command() {
	case "save":
		if isUserAdmin(bot, update.Message) {

			err := saveMessage(update.Message)
			if err != nil {
				msg.Text = "Ошибка в сохранении сообщения."
			} else {
				msg.Text = "Сообщение успешно сохранено."
			}
		} else {
			msg.Text = "Вы должны быть администратором, чтобы выполнить это действие."
		}
	case "status":
		msg.Text = fmt.Sprintf("Бот онлайн. Последняя проверка в %s", time.Now().Format(time.RFC1123))
	default:
		msg.Text = "Я не знаю эту команду."
	}

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func isUserAdmin(bot *tgbotapi.BotAPI, message *tgbotapi.Message) bool {
	chatMemberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: message.Chat.ID,
			UserID: message.From.ID,
		},
	}

	chatMember, err := bot.GetChatMember(chatMemberConfig)
	if err != nil {
		log.Print(err)
		return false
	}

	return chatMember.Status == "administrator" || chatMember.Status == "creator"
}

func saveMessage(message *tgbotapi.Message) error {
	// Save the message to a text file
	fileName := fmt.Sprintf("messages_%d.txt", message.Chat.ID)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(fmt.Sprintf("%s: %s\n", message.From.UserName, message.Text)); err != nil {
		return err
	}
	return nil
}

// функция для проверки бот в группе или нет, необходимо попросить бота взять из группы последнее сообщение, если он не сможет прочитать сообщеине, то он оффлайн
