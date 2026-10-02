package tasks

import (
	"famoria/internal/bot/idle/item"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type MiningOpts struct {
	Log          *zap.Logger
	BrakRepo     brak.Repository
	ChatSettings chat_settings.Repository
	Manager      *item.Manager
	S            gocron.Scheduler
}

func StartMining(opts MiningOpts) {
	_, err := opts.S.NewJob(
		gocron.DurationJob(
			2*time.Hour,
		),
		gocron.NewTask(
			func(opts MiningOpts) {
				braks, err := opts.BrakRepo.FindAllMining()
				if err != nil {
					return
				}
				for _, b := range braks {
					// A chat that switched earnings off must stop accumulating
					// score, including from this scheduled job. Braks without a
					// recorded chat keep mining, as before: there is no chat
					// setting to honour.
					if b.ChatID != 0 && !opts.ChatSettings.IsFeatureEnabled(b.ChatID, chat_settings.FeatEarnings) {
						opts.Log.Info("[CRON] brak id: " + b.OID.Hex() + " skipped, earnings disabled in chat")
						continue
					}
					b.ApplyBuffs(opts.Manager)
					resp := b.Events.Mining.Play()
					if resp.Score > 0 {
						b.Events.Mining.LastPlay = time.Now()
						opts.Log.Info("[CRON] brak id: " + b.OID.Hex() + "luck mining")
					} else if resp.Score != 0 {
						opts.Log.Info("[CRON] brak id: " + b.OID.Hex() + "unluck sgorel mining")
					} else {
						opts.Log.Info("[CRON] brak id: " + b.OID.Hex() + "unluck mining")
						continue
					}
					err = opts.BrakRepo.Update(
						bson.M{"_id": b.OID},
						bson.M{
							"$inc": bson.M{
								"score": resp.Score,
							},
							"$set": bson.M{
								"events.mining": b.Events.Mining,
							},
						},
					)
					if err != nil {
						opts.Log.Sugar().Error("Ошибка при обновлении счёта #mining (", resp.Score, ") брака ", b.OID, ":", err)
						continue
					}
				}
			},
			opts,
		),
	)
	if err != nil {
		opts.Log.Sugar().Error("Ошибка при создании задачи cron для mining:", err)
		panic(err)
	}
}
