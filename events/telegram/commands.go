package telegram

import (
	"BotSavingPages/clients/telegram"
	"context"
	"errors"
	"net/url"
	"strings"

	"BotSavingPages/lib/e"
	"BotSavingPages/storage"
)

const (
	StartCmd = "/start"
	CmdHelp  = "Help"
	CmdGet   = "Get"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case StartCmd:
		return p.sendHelloWithButtons(chatID)
	case CmdHelp:
		return p.sendHelp(chatID)
	case CmdGet:
		return p.sendRandom(chatID, username)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) sendHelloWithButtons(chatID int) error {
	inlineKeyboard := [][]telegram.KeyboardButton{
		{
			{Text: "Help"},
			{Text: "Get"},
		},
	}

	return p.tg.SendMessageWithReplyKeyboard(chatID, msgHello, inlineKeyboard)
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(context.Background(), page)
	if err != nil {
		return err
	}
	if isExists {
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(context.Background(), page)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
