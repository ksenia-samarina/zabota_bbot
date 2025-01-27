package clients

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksenia-samarina/zabota_bbot/internal/logger"
	"github.com/ksenia-samarina/zabota_bbot/internal/model"
	"github.com/pkg/errors"
	"strings"
)

type HandlerFunc func(tgUpdate tgbotapi.Update, c *Client, msgModel *model.Model)

func (f HandlerFunc) RunFunc(tgUpdate tgbotapi.Update, c *Client, msgModel *model.Model) {
	f(tgUpdate, c, msgModel)
}

type Client struct {
	client                *tgbotapi.BotAPI
	handlerProcessingFunc HandlerFunc
}

type TokenGetter interface {
	Token() string
}

func New(tokenGetter TokenGetter, handlerProcessingFunc HandlerFunc) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(tokenGetter.Token())
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка NewBotAPI")
	}

	return &Client{
		client:                client,
		handlerProcessingFunc: handlerProcessingFunc,
	}, nil
}

func (c *Client) SendMessage(text string, userID int64) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = "markdown"
	_, err := c.client.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Ошибка отправки сообщения client.Send")
	}
	return nil
}

func (c *Client) ListenUpdates(msgModel *model.Model) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.client.GetUpdatesChan(u)

	logger.Info("Start listening for tg messages")

	for update := range updates {
		c.handlerProcessingFunc.RunFunc(update, c, msgModel)
	}
}

func ProcessingMessages(tgUpdate tgbotapi.Update, c *Client, msgModel *model.Model) {
	if tgUpdate.Message != nil {
		logger.Info(fmt.Sprintf("[%s][%v] %s", tgUpdate.Message.From.UserName, tgUpdate.Message.From.ID, tgUpdate.Message.Text))
		err := msgModel.IncomingMessage(&model.Message{
			Text:            tgUpdate.Message.Text,
			UserID:          tgUpdate.Message.From.ID,
			UserName:        tgUpdate.Message.From.UserName,
			UserDisplayName: strings.TrimSpace(tgUpdate.Message.From.FirstName + " " + tgUpdate.Message.From.LastName),
		})
		if err != nil {
			logger.Error("error processing message:", "err", err)
		}
	} else if tgUpdate.CallbackQuery != nil {
		logger.Info(fmt.Sprintf("[%s][%v] Callback: %s", tgUpdate.CallbackQuery.From.UserName, tgUpdate.CallbackQuery.From.ID, tgUpdate.CallbackQuery.Data))
		callback := tgbotapi.NewCallback(tgUpdate.CallbackQuery.ID, tgUpdate.CallbackQuery.Data)
		if _, err := c.client.Request(callback); err != nil {
			logger.Error("Ошибка Request callback:", "err", err)
		}
		if err := deleteInlineButtons(c, tgUpdate.CallbackQuery.From.ID, tgUpdate.CallbackQuery.Message.MessageID, tgUpdate.CallbackQuery.Message.Text); err != nil {
			logger.Error("Ошибка удаления кнопок:", "err", err)
		}
		err := msgModel.IncomingMessage(&model.Message{
			Text:            tgUpdate.CallbackQuery.Data,
			UserID:          tgUpdate.CallbackQuery.From.ID,
			UserName:        tgUpdate.CallbackQuery.From.UserName,
			UserDisplayName: strings.TrimSpace(tgUpdate.CallbackQuery.From.FirstName + " " + tgUpdate.CallbackQuery.From.LastName),
			IsCallback:      true,
			CallbackMsgID:   tgUpdate.CallbackQuery.InlineMessageID,
		})
		if err != nil {
			logger.Error("error processing message from callback:", "err", err)
		}
	}
}

func (c *Client) ShowInlineButtons(text string, buttons []model.TgRowButtons, userID int64) error {
	keyboard := make([][]tgbotapi.InlineKeyboardButton, len(buttons))
	for i := 0; i < len(buttons); i++ {
		tgRowButtons := buttons[i]
		keyboard[i] = make([]tgbotapi.InlineKeyboardButton, len(tgRowButtons))
		for j := 0; j < len(tgRowButtons); j++ {
			tgInlineButton := tgRowButtons[j]
			keyboard[i][j] = tgbotapi.NewInlineKeyboardButtonData(tgInlineButton.DisplayName, tgInlineButton.Value)
		}
	}
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = numericKeyboard
	msg.ParseMode = "markdown"
	_, err := c.client.Send(msg)
	if err != nil {
		logger.Error("Ошибка отправки сообщения", "err", err)
		return errors.Wrap(err, "client.Send with inline-buttons")
	}
	return nil
}

func deleteInlineButtons(c *Client, userID int64, msgID int, sourceText string) error {
	msg := tgbotapi.NewEditMessageText(userID, msgID, sourceText)
	_, err := c.client.Send(msg)
	if err != nil {
		logger.Error("Ошибка отправки сообщения", "err", err)
		return errors.Wrap(err, "client.Send remove inline-buttons")
	}
	return nil
}
