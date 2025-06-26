package main

import (
	"BotSavingPages/storage/sqlite"
	"context"
	"flag"
	"log"

	tgClient "BotSavingPages/clients/telegram"
	"BotSavingPages/consumer/event-consumer"
	"BotSavingPages/events/telegram"
)

const (
	tgBotHost         = "api.telegram.org"
	sqliteStoragePath = "data/sqlite"
	batchSize         = 100
)

func main() {
	s, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatal("can't connect to sqlite storage", err)
	}

	if err := s.Init(context.TODO()); err != nil {
		log.Fatal("can't init sqlite storage", err)
	}

	botToken := "7286014980:AAHlKmajCsdslVA-KQZCFPYV__cGlAadk50"
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, botToken),
		s,
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}

func mustToken() string {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token
}
