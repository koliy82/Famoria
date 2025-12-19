package farm_logger

import (
	"context"
	"errors"
	"famoria/internal/bot/callback"
	"famoria/internal/config"
	"famoria/internal/database/steamapi/repositories/steam_accounts"
	"famoria/internal/pkg/common"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type FarmLogger struct {
	db     *mongo.Database
	coll   *mongo.Collection
	log    *zap.Logger
	cfg    config.Config
	ctx    context.Context
	cancel context.CancelFunc
	bot    *telego.Bot
	cm     *callback.CallbacksManager
	api    *steam_accounts.SteamAPI
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config, bot *telego.Bot, cm *callback.CallbacksManager, api *steam_accounts.SteamAPI) *FarmLogger {
	if cfg.MongoFarmLogsCollName == nil || *cfg.MongoFarmLogsCollName == "" {
		log.Warn("MONGO_FARM_LOGS_COLL_NAME is nil, skipping farm logs repository initialization")
		return nil
	}
	db := client.Database(*cfg.MongoSteamDatabase)
	coll := db.Collection(*cfg.MongoFarmLogsCollName)
	ctx, cancel := context.WithCancel(context.Background())
	fl := &FarmLogger{
		db:     db,
		coll:   coll,
		log:    log,
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		bot:    bot,
		cm:     cm,
		api:    api,
	}
	fl.init()
	return fl
}

func (l *FarmLogger) init() {
	resumeToken, err := l.loadResumeToken()
	if err != nil {
		l.log.Error("Error loading resume token", zap.Error(err))
	}
	streamOptions := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	if len(resumeToken) > 0 {
		l.log.Info("Resuming change stream from saved token")
		streamOptions.SetResumeAfter(resumeToken)
	} else {
		l.log.Info("Starting new change stream")
	}
	go l.startWaiter(streamOptions)
}

func (l *FarmLogger) startWaiter(streamOptions *options.ChangeStreamOptions) {
	pipeline := mongo.Pipeline{bson.D{{"$match", bson.D{{"operationType", "insert"}}}}}
	cs, err := l.coll.Watch(l.ctx, pipeline, streamOptions)
	if err != nil {
		l.log.Error("Error starting change stream", zap.Error(err))
		return
	}
	defer func() {
		if err := cs.Close(l.ctx); err != nil {
			l.log.Error("Error closing change stream", zap.Error(err))
		}
	}()
	l.log.Debug("Starting farm logs watcher")

	for cs.Next(l.ctx) {
		var event struct {
			FullDocument FarmLog `bson:"fullDocument"`
		}
		if err := cs.Decode(&event); err != nil {
			l.log.Error("Error decoding change event", zap.Error(err))
			continue
		}

		fl := event.FullDocument
		l.handleFarmLog(fl)

		rt := cs.ResumeToken()
		if len(rt) > 0 {
			if err := l.saveResumeToken(rt); err != nil {
				l.log.Error("Error saving resume token", zap.Error(err))
			}
		} else {
			l.log.Warn("Warning: No resume token available to save")
		}
	}

	if err := cs.Err(); err != nil {
		l.log.Error("Change stream error", zap.Error(err))
	}
}

func (l *FarmLogger) handleFarmLog(fl FarmLog) {
	params := &telego.SendMessageParams{
		ChatID:              tu.ID(fl.TelegramID),
		ParseMode:           "",
		DisableNotification: false,
	}
	isFarming := fl.Reason == GamesSend
	switch fl.Reason {
	case GamesSend, UserStop:
		Callback := l.cm.DynamicCallback(callback.DynamicOpts{
			Label:    common.Ternary(isFarming, "Остановить фарм ⏸️", "Продолжить фарм ▶️"),
			CtxType:  callback.OneClick,
			OwnerIDs: []int64{fl.TelegramID},
			Time:     time.Duration(24) * time.Hour,
			Callback: func(query telego.CallbackQuery) {
				var err error
				if isFarming {
					err = l.api.StopFarming(fl.SteamID)
				} else {
					err = l.api.StartFarming(fl.SteamID)
				}
				if err != nil {
					l.log.Error("failed to edit update status message text", zap.Error(err))
					return
				}
			},
		})
		params.WithText("Фарм для аккаунта " + fl.SteamUsername() + common.Ternary(isFarming, " начат.", " остановлен."))
		params.WithReplyMarkup(tu.InlineKeyboard(tu.InlineKeyboardRow(Callback.Inline())))
	case UserDelete:
		params.WithText("Аккаунт " + fl.SteamUsername() + " успешно удалён из бота.")
	default:
		panic("unhandled default case")
	}
	_, err := l.bot.SendMessage(context.Background(), params)
	if err != nil {
		l.log.Error("Error sending steam log message", zap.Error(err))
	}
}

func (l *FarmLogger) saveResumeToken(token bson.Raw) error {
	metadataColl := l.db.Collection("metadataColl")

	l.log.Debug("Saving resume token", zap.Any("token", token))

	_, err := metadataColl.UpdateOne(
		context.Background(),
		bson.D{{Key: "_id", Value: "steam_farm_token"}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "tokenBinary", Value: token}, {Key: "updatedAt", Value: time.Now()}}}},
		options.Update().SetUpsert(true),
	)

	return err
}

func (l *FarmLogger) loadResumeToken() (bson.Raw, error) {
	metadataColl := l.db.Collection("metadataColl")

	var result struct {
		ID          string    `bson:"_id"`
		TokenBinary bson.Raw  `bson:"tokenBinary"`
		UpdatedAt   time.Time `bson:"updatedAt"`
	}

	err := metadataColl.FindOne(context.Background(), bson.D{{Key: "_id", Value: "steam_farm_token"}}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			l.log.Info("No resume token found in database")
			return nil, nil
		}
		return nil, err
	}

	if len(result.TokenBinary) == 0 {
		l.log.Info("Resume token is empty")
		return nil, nil
	}

	l.log.Debug("Loaded resume token", zap.Any("token", result.TokenBinary), zap.Time("updatedAt", result.UpdatedAt))
	return result.TokenBinary, nil
}
