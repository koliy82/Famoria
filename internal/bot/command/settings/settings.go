package settings

import (
	"context"
	"time"

	"famoria/internal/bot/callback"
	"famoria/internal/bot/predicate"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// buttonLifetime bounds how long a settings button stays clickable. Navigating
// re-renders the message with fresh buttons, so stale ones simply expire.
const buttonLifetime = 30 * time.Minute

// screen identifies which settings page is shown in the message.
type screen int

const (
	screenRoot screen = iota
	screenLang
	screenVideo
	screenMarriage
	screenGacha
)

// screenForFeature maps a toggleable feature to the category page that owns it,
// so a toggle re-renders the page the user is actually on.
func screenForFeature(f chat_settings.Feature) screen {
	for _, g := range chat_settings.GachaFeatures() {
		if g == f {
			return screenGacha
		}
	}
	return screenMarriage
}

type Opts struct {
	fx.In
	Bh           *th.BotHandler
	Log          *zap.Logger
	Bot          *telego.Bot
	Cm           *callback.CallbacksManager
	ChatSettings chat_settings.Repository
}

type manager struct {
	log  *zap.Logger
	bot  *telego.Bot
	cm   *callback.CallbacksManager
	repo chat_settings.Repository
}

// session is one settings message. It carries the identifiers needed to edit
// that message back, and re-reads the current state from the repository on every
// render so the page always reflects what was just stored.
type session struct {
	mgr       *manager
	chatID    int64
	userID    int64
	messageID int
	isPrivate bool
}

func Register(opts Opts) {
	m := &manager{
		log:  opts.Log,
		bot:  opts.Bot,
		cm:   opts.Cm,
		repo: opts.ChatSettings,
	}

	opts.Bh.Handle(m.handleSettings, th.CommandEqual("settings"))
}

// canConfigure reports whether the sender may change this chat's settings.
//
// In a private chat the user owns the conversation, so no administrator check
// applies — and GetChatMember never reports "administrator" there, which would
// otherwise lock private chats out of their own settings.
func (m *manager) canConfigure(msg *telego.Message) bool {
	if msg.Chat.Type == telego.ChatTypePrivate {
		return true
	}
	return predicate.IsChatAdmin(context.Background(), m.bot, msg.Chat.ID, msg.From.ID)
}

// handleSettings is the /settings entry point. Only administrators of the chat
// the command was sent in may configure it.
func (m *manager) handleSettings(ctx *th.Context, update telego.Update) error {
	msg := update.Message
	if msg == nil || msg.From == nil {
		return nil
	}
	chatID := msg.Chat.ID
	lang := m.repo.Lang(chatID)

	if !m.canConfigure(msg) {
		_, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID:    tu.ID(chatID),
			ParseMode: telego.ModeHTML,
			Text:      i18n.T(lang, i18n.KeySettingsAdmin),
			ReplyParameters: &telego.ReplyParameters{
				MessageID:                msg.MessageID,
				AllowSendingWithoutReply: true,
			},
			DisableNotification: true,
		})
		if err != nil {
			m.log.Sugar().Error(err)
		}
		return err
	}

	s := &session{mgr: m, chatID: chatID, userID: msg.From.ID, isPrivate: msg.Chat.Type == telego.ChatTypePrivate}

	text, kb := s.render(screenRoot)
	sent, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		ParseMode:   telego.ModeHTML,
		Text:        text,
		ReplyMarkup: kb,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                msg.MessageID,
			AllowSendingWithoutReply: true,
		},
		DisableNotification: true,
	})
	if err != nil {
		m.log.Sugar().Error(err)
		return err
	}
	s.messageID = sent.MessageID
	return nil
}

// lang is read fresh on every call so switching the language immediately
// re-renders the rest of the menu in it.
func (s *session) lang() i18n.Lang { return s.mgr.repo.Lang(s.chatID) }

func (s *session) t(key string, args ...any) string {
	return i18n.T(s.lang(), key, args...)
}

// authorized re-checks rights when a button is pressed, so losing chat admin
// rights also loses the ability to keep changing settings. The button's owner
// is matched as well, since only that user should drive the menu they opened.
func (s *session) authorized(q telego.CallbackQuery) bool {
	if q.From.ID == 0 || q.From.ID != s.userID {
		return false
	}
	if s.isPrivate {
		return true
	}
	return predicate.IsChatAdmin(context.Background(), s.mgr.bot, s.chatID, q.From.ID)
}

// show replaces the settings message content with the given screen.
func (s *session) show(sc screen) {
	text, kb := s.render(sc)
	_, err := s.mgr.bot.EditMessageText(context.Background(), &telego.EditMessageTextParams{
		ChatID:      tu.ID(s.chatID),
		MessageID:   s.messageID,
		ParseMode:   telego.ModeHTML,
		Text:        text,
		ReplyMarkup: kb,
	})
	if err != nil {
		s.mgr.log.Sugar().Error(err)
	}
}

// navButton builds a button that switches the message to another screen.
func (s *session) navButton(label string, sc screen) telego.InlineKeyboardButton {
	cb := s.mgr.cm.DynamicCallback(callback.DynamicOpts{
		Label:    label,
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.userID},
		Time:     buttonLifetime,
		Callback: func(q telego.CallbackQuery) {
			if !s.authorized(q) {
				return
			}
			s.show(sc)
		},
	})
	return cb.Inline()
}

// stateIcon prefixes a toggle label with its current state.
func stateIcon(enabled bool) string {
	if enabled {
		return "✅"
	}
	return "❌"
}

// featureToggle builds a button flipping one chat feature, then re-rendering the
// owning category page so the new state is visible.
func (s *session) featureToggle(f chat_settings.Feature) telego.InlineKeyboardButton {
	label := stateIcon(s.mgr.repo.IsFeatureEnabled(s.chatID, f)) + " " + f.Label(s.lang())

	cb := s.mgr.cm.DynamicCallback(callback.DynamicOpts{
		Label:      label,
		CtxType:    callback.Temporary,
		OwnerIDs:   []int64{s.userID},
		Time:       buttonLifetime,
		AnswerText: s.t(i18n.KeySettingsUpdated),
		Callback: func(q telego.CallbackQuery) {
			if !s.authorized(q) {
				return
			}
			next := !s.mgr.repo.IsFeatureEnabled(s.chatID, f)
			s.mgr.repo.SetFeatureEnabled(s.chatID, f, next)
			s.show(screenForFeature(f))
		},
	})
	return cb.Inline()
}

// backRow is the shared "return to categories" row.
func (s *session) backRow() []telego.InlineKeyboardButton {
	return tu.InlineKeyboardRow(s.navButton(s.t(i18n.KeySettingsBack), screenRoot))
}

// render builds the text and keyboard for a screen.
func (s *session) render(sc screen) (string, *telego.InlineKeyboardMarkup) {
	switch sc {
	case screenLang:
		return s.renderLang()
	case screenVideo:
		return s.renderVideo()
	case screenMarriage:
		return s.renderMarriage()
	case screenGacha:
		return s.renderGacha()
	default:
		return s.renderRoot()
	}
}

func (s *session) renderRoot() (string, *telego.InlineKeyboardMarkup) {
	text := s.t(i18n.KeySettingsTitle) + "\n\n" + s.t(i18n.KeySettingsDesc)

	kb := tu.InlineKeyboard(
		tu.InlineKeyboardRow(s.navButton(s.t(i18n.KeyCatLang), screenLang)),
		tu.InlineKeyboardRow(s.navButton(s.t(i18n.KeyCatVideo), screenVideo)),
		tu.InlineKeyboardRow(s.navButton(s.t(i18n.KeyCatMarriage), screenMarriage)),
		tu.InlineKeyboardRow(s.navButton(s.t(i18n.KeyCatGacha), screenGacha)),
	)
	return text, kb
}

func (s *session) renderLang() (string, *telego.InlineKeyboardMarkup) {
	current := s.lang()
	text := s.t(i18n.KeySettingsLangTitle) + "\n\n" +
		s.t(i18n.KeySettingsLangDesc, i18n.LangName(current))

	rows := make([][]telego.InlineKeyboardButton, 0, 2)
	for _, l := range i18n.SupportedLangs {
		label := i18n.LangName(l)
		if l == current {
			label = "✓ " + label
		}
		rows = append(rows, tu.InlineKeyboardRow(s.langButton(label, l)))
	}
	rows = append(rows, s.backRow())

	return text, tu.InlineKeyboard(rows...)
}

// langButton stores the chosen language and re-renders the page in it, so the
// user immediately sees the rest of the menu translated.
func (s *session) langButton(label string, l i18n.Lang) telego.InlineKeyboardButton {
	cb := s.mgr.cm.DynamicCallback(callback.DynamicOpts{
		Label:      label,
		CtxType:    callback.Temporary,
		OwnerIDs:   []int64{s.userID},
		Time:       buttonLifetime,
		AnswerText: s.t(i18n.KeySettingsUpdated),
		Callback: func(q telego.CallbackQuery) {
			if !s.authorized(q) {
				return
			}
			s.mgr.repo.SetLang(s.chatID, l)
			s.show(screenLang)
		},
	})
	return cb.Inline()
}

func (s *session) renderVideo() (string, *telego.InlineKeyboardMarkup) {
	text := s.t(i18n.KeySettingsVideoTitle) + "\n\n" + s.t(i18n.KeySettingsVideoDesc)

	enabled := s.mgr.repo.IsVideoConverterEnabled(s.chatID)
	keep := !s.mgr.repo.DeleteOriginalLink(s.chatID)

	kb := tu.InlineKeyboard(
		tu.InlineKeyboardRow(s.videoConverterButton(enabled)),
		tu.InlineKeyboardRow(s.videoDeleteButton(keep)),
		s.backRow(),
	)
	return text, kb
}

func (s *session) videoConverterButton(enabled bool) telego.InlineKeyboardButton {
	key := i18n.KeyVideoConverterOff
	if enabled {
		key = i18n.KeyVideoConverterOn
	}
	cb := s.mgr.cm.DynamicCallback(callback.DynamicOpts{
		Label:      s.t(key),
		CtxType:    callback.Temporary,
		OwnerIDs:   []int64{s.userID},
		Time:       buttonLifetime,
		AnswerText: s.t(i18n.KeySettingsUpdated),
		Callback: func(q telego.CallbackQuery) {
			if !s.authorized(q) {
				return
			}
			s.mgr.repo.SetVideoConverter(s.chatID, !enabled)
			s.show(screenVideo)
		},
	})
	return cb.Inline()
}

func (s *session) videoDeleteButton(keep bool) telego.InlineKeyboardButton {
	key := i18n.KeyVideoDeleteOn
	if keep {
		key = i18n.KeyVideoDeleteOff
	}
	cb := s.mgr.cm.DynamicCallback(callback.DynamicOpts{
		Label:      s.t(key),
		CtxType:    callback.Temporary,
		OwnerIDs:   []int64{s.userID},
		Time:       buttonLifetime,
		AnswerText: s.t(i18n.KeySettingsUpdated),
		Callback: func(q telego.CallbackQuery) {
			if !s.authorized(q) {
				return
			}
			s.mgr.repo.SetVideoKeepOriginal(s.chatID, !keep)
			s.show(screenVideo)
		},
	})
	return cb.Inline()
}

func (s *session) renderMarriage() (string, *telego.InlineKeyboardMarkup) {
	text := s.t(i18n.KeySettingsBrakTitle) + "\n\n" + s.t(i18n.KeySettingsBrakDesc)

	features := chat_settings.MarriageFeatures()
	rows := make([][]telego.InlineKeyboardButton, 0, len(features)+1)
	for _, f := range features {
		rows = append(rows, tu.InlineKeyboardRow(s.featureToggle(f)))
	}
	rows = append(rows, s.backRow())

	return text, tu.InlineKeyboard(rows...)
}

func (s *session) renderGacha() (string, *telego.InlineKeyboardMarkup) {
	text := s.t(i18n.KeySettingsGachaTitle) + "\n\n" + s.t(i18n.KeySettingsGachaDesc)
	text += "\n\n" + s.t(i18n.KeySettingsEarnWarning)

	features := chat_settings.GachaFeatures()
	rows := make([][]telego.InlineKeyboardButton, 0, len(features)+1)
	for _, f := range features {
		rows = append(rows, tu.InlineKeyboardRow(s.featureToggle(f)))
	}
	rows = append(rows, s.backRow())

	return text, tu.InlineKeyboard(rows...)
}
