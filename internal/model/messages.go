package model

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"strings"
	"time"
)

var btnStart = []TgRowButtons{
	{TgInlineButton{DisplayName: "Добавить анкету", Value: "/add_form"}, TgInlineButton{DisplayName: "Изменить анкету", Value: "/update_form"}},
	{TgInlineButton{DisplayName: "Задать вопрос", Value: "/ask_question"}},
}

type MessageSender interface {
	SendMessage(text string, userID int64) error
	ShowInlineButtons(text string, buttons []TgRowButtons, userID int64) error
}

type UserDataStorage interface {
	InsertUserDataRecord(ctx context.Context, userID int64, rec UserDataRecord, userName string) error
	UpdateUserDataRecord(ctx context.Context, userID int64, rec UserDataRecord, userName string) error
	GetUserDataRecord(ctx context.Context, userID int64, period time.Time) ([]UserDataFormRecord, error)
}

type AIModule interface {
	AnswerQuestion(question string, form string) (string, error)
}

type cache interface {
	Get(key string) any
	Put(key string, value any)
	Delete(key string)
}

type kafkaProducer interface {
	SendMessage(key string, value string) (partition int32, offset int64, err error)
	GetTopic() string
}

type Model struct {
	ctx             context.Context
	tgClient        MessageSender
	storage         UserDataStorage
	kafkaProducer   kafkaProducer
	cache           cache
	lastUserCommand map[int64]string
}

func New(ctx context.Context, tgClient MessageSender) *Model {
	return &Model{
		ctx:             ctx,
		tgClient:        tgClient,
		lastUserCommand: make(map[int64]string),
	}
}

type Message struct {
	Text            string
	UserID          int64
	UserName        string
	UserDisplayName string
	IsCallback      bool
	CallbackMsgID   string
}

func (s *Model) GetCtx() context.Context {
	return s.ctx
}

func (s *Model) SetCtx(ctx context.Context) {
	s.ctx = ctx
}

func (s *Model) IncomingMessage(msg *Message) error {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "IncomingMessage")
	s.ctx = ctx
	defer span.Finish()

	lastUserCommand := s.lastUserCommand[msg.UserID]
	s.lastUserCommand[msg.UserID] = ""

	if isNeedReturn, err := checkIfEnterNewForm(s, msg, lastUserCommand); err != nil || isNeedReturn {
		return err
	}
	if isNeedReturn, err := checkIfUpdateCurForm(s, msg, lastUserCommand); err != nil || isNeedReturn {
		return err
	}
	if isNeedReturn, err := checkIfAskQuestion(s, msg, lastUserCommand); err != nil || isNeedReturn {
		return err
	}
	if isNeedReturn, err := checkIfChoiceEnterNewForm(s, msg); err != nil || isNeedReturn {
		return err
	}
	if isNeedReturn, err := checkIfChoiceUpdateCurForm(s, msg); err != nil || isNeedReturn {
		return err
	}
	if isNeedReturn, err := checkIfChoiceAskQuestion(s, msg); err != nil || isNeedReturn {
		return err
	}
	if isNeedReturn, err := checkBotCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	return s.tgClient.SendMessage(txtUnknownCommand, msg.UserID)
}

func checkIfAskQuestion(s *Model, msg *Message, lastUserCommand string) (bool, error) {
	if lastUserCommand == "/ask_question" {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfAskQuestion")
		s.ctx = ctx
		defer span.Finish()

		return true, s.tgClient.SendMessage(txtQuestionAsked, msg.UserID)
	}
	return false, nil
}

func checkIfEnterNewForm(s *Model, msg *Message, lastUserCommand string) (bool, error) {
	if lastUserCommand == "/add_form" {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfEnterNewForm")
		s.ctx = ctx
		defer span.Finish()

		//userDataRecord := UserDataRecord{
		//	UserID: msg.UserID,
		//	Form:   msg.Text,
		//}
		//err := s.storage.InsertUserDataRecord(s.ctx, msg.UserID, userDataRecord, msg.UserName)
		//if err != nil {
		//	logger.Error("Ошибка сохранения анкеты о ребенке", "err", err)
		//	return true, errors.Wrap(err, "Insert form error")
		//}
		return true, s.tgClient.SendMessage(txtFormAdded, msg.UserID)
	}
	return false, nil
}

func checkIfUpdateCurForm(s *Model, msg *Message, lastUserCommand string) (bool, error) {
	if lastUserCommand == "/update_form" {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfUpdateCurForm")
		s.ctx = ctx
		defer span.Finish()

		//userDataRecord := UserDataRecord{
		//	UserID: msg.UserID,
		//	Form:   msg.Text,
		//}
		//err := s.storage.UpdateUserDataRecord(s.ctx, msg.UserID, userDataRecord, msg.UserName)
		//if err != nil {
		//	logger.Error("Ошибка обновления анкеты о ребенке", "err", err)
		//	return true, errors.Wrap(err, "Update form error")
		//}
		return true, s.tgClient.SendMessage(txtFormUpdated, msg.UserID)
	}
	return false, nil
}

func checkIfChoiceEnterNewForm(s *Model, msg *Message) (bool, error) {
	if msg.IsCallback {
		if strings.Contains(msg.Text, "/add_form") {
			span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfChoiceEnterNewForm")
			s.ctx = ctx
			defer span.Finish()

			s.lastUserCommand[msg.UserID] = "/add_form"
			return true, s.tgClient.SendMessage(txtFormAdd, msg.UserID)
		}
	}
	return false, nil
}

func checkIfChoiceUpdateCurForm(s *Model, msg *Message) (bool, error) {
	if msg.IsCallback {
		if strings.Contains(msg.Text, "/update_form") {
			span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfChoiceUpdateCurForm")
			s.ctx = ctx
			defer span.Finish()

			s.lastUserCommand[msg.UserID] = "/update_form"
			return true, s.tgClient.SendMessage(txtFormUpdate, msg.UserID)
		}
	}
	return false, nil
}

func checkIfChoiceAskQuestion(s *Model, msg *Message) (bool, error) {
	if msg.IsCallback {
		if strings.Contains(msg.Text, "/ask_question") {
			span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfChoiceAskQuestion")
			s.ctx = ctx
			defer span.Finish()

			s.lastUserCommand[msg.UserID] = "/ask_question"
			return true, s.tgClient.SendMessage(txtQuestionAsk, msg.UserID)
		}
	}
	return false, nil
}

func checkBotCommands(s *Model, msg *Message) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkBotCommands")
	s.ctx = ctx
	defer span.Finish()

	switch msg.Text {
	case "/start":
		displayName := msg.UserDisplayName
		if len(displayName) == 0 {
			displayName = msg.UserName
		}
		return true, s.tgClient.ShowInlineButtons(fmt.Sprintf(txtStart, displayName), btnStart, msg.UserID)
	case "/help":
		return true, s.tgClient.SendMessage(txtHelp, msg.UserID)
	}
	return false, nil
}
