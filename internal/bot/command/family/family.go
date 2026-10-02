package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/predicate"
	"famoria/internal/config"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/database/mongo/repositories/message"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/i18n"

	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh           *th.BotHandler
	Log          *zap.Logger
	Cfg          config.Config
	BrakRepo     brak.Repository
	UserRepo     user.Repository
	MessageRepo  message.Repository
	ChatSettings chat_settings.Repository
	Cm           *callback.CallbacksManager
}

func Register(opts Opts) {
	// Each command is gated on its own chat-level feature switch: a chat whose
	// administrators turned a feature off gets no response for it at all.

	opts.Bh.Handle(profileCmd{
		cm:           opts.Cm,
		brakRepo:     opts.BrakRepo,
		userRepo:     opts.UserRepo,
		messageRepo:  opts.MessageRepo,
		chatSettings: opts.ChatSettings,
		log:          opts.Log,
	}.Handle, th.Or(
		th.CommandEqual("profile"),
		predicate.TextEqualKey(i18n.KeyBtnProfile),
		th.CommandEqual("mybrak"),
	))

	opts.Bh.Handle(goFamilyCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.CommandEqual("gobrak"),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatGoBrak),
	))

	opts.Bh.Handle(endFamilyCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("endbrak"), predicate.TextEqualKey(i18n.KeyBtnDivorce)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatEndBrak),
	))

	opts.Bh.Handle(goKidCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.CommandEqual("kid"),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatKid),
	))

	opts.Bh.Handle(endKidCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("kidannihilate"), predicate.TextEqualKey(i18n.KeyBtnAnnihilate)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatKidAnnihilate),
	))

	opts.Bh.Handle(leaveKidCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("detdom"), predicate.TextEqualKey(i18n.KeyBtnOrphanage)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatDetdom),
	))

	opts.Bh.Handle(pagesCmd{
		cm:           opts.Cm,
		brakRepo:     opts.BrakRepo,
		chatSettings: opts.ChatSettings,
		isLocal:      true,
		log:          opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("braks"), predicate.TextEqualKey(i18n.KeyFeatureBraks)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatBraks),
	))

	opts.Bh.Handle(pagesCmd{
		cm:           opts.Cm,
		brakRepo:     opts.BrakRepo,
		chatSettings: opts.ChatSettings,
		isLocal:      false,
		log:          opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("braksglobal"), predicate.TextEqualKey(i18n.KeyFeatureBraksGlobal)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatBraksGlobal),
	))

	opts.Bh.Handle(treeCmd{
		cfg: opts.Cfg,
		log: opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("tree"), predicate.TextEqualKey(i18n.KeyBtnTree)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatTree),
	))

	// Deposits and withdrawals move the marriage balance, which is part of chat
	// earnings; they follow the same switch as the job buttons in the profile.
	opts.Bh.Handle(depositCmd{
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqualArgc("deposit", 1), th.CommandEqualArgc("dep", 1)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatEarnings),
	))

	opts.Bh.Handle(withdrawCmd{
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqualArgc("withdraw", 1), th.CommandEqualArgc("with", 1)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatEarnings),
	))
}
