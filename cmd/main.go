package main

import (
	"context"
	"github.com/ksenia-samarina/zabota_bbot/internal/clients"
	"github.com/ksenia-samarina/zabota_bbot/internal/config"
	"github.com/ksenia-samarina/zabota_bbot/internal/logger"
	"github.com/ksenia-samarina/zabota_bbot/internal/model"
)

func main() {

	logger.Info("Старт приложения")

	ctx := context.Background()

	config, err := config.New()
	if err != nil {
		logger.Fatal("Ошибка получения файла конфигурации:", "err", err)
	}

	// Оборачивание в Middleware функции обработки сообщения для метрик и трейсинга.
	tgProcessingFuncHandler := clients.HandlerFunc(clients.ProcessingMessages)

	// Инициализация телеграм клиента.
	tgClient, err := clients.New(config, tgProcessingFuncHandler)
	if err != nil {
		logger.Fatal("Ошибка инициализации ТГ-клиента:", "err", err)
	}

	// Инициализация основной модели.
	msgModel := model.New(ctx, tgClient)

	// Запуск ТГ-клиента.
	tgClient.ListenUpdates(msgModel)

	logger.Info("Завершение приложения")
}
