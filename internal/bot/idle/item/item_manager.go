package item

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/events/hamster"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/pkg/common"
	"go.uber.org/zap"
)

type Manager struct {
	Log   *zap.Logger
	Items map[items.Name]*Item
}

func (i *Manager) GetItem(name items.Name) *Item {
	item := i.Items[name]
	if item == nil {
		i.Log.Sugar().Error("Item not found", name)
	}
	return item
}

type Item struct {
	Name        items.Name
	Emoji       string
	MaxLevel    int
	Buffs       map[int][]events.Buff
	Description string
	Prices      map[int]*common.Score
}

func New(log *zap.Logger) *Manager {
	return &Manager{
		Log: log,
		Items: map[items.Name]*Item{
			items.MegaTap: {
				Emoji:       "💪",
				Name:        items.MegaTap,
				MaxLevel:    5,
				Description: "Священная перчатка, усиливающая силу вашего тапа.",
				Buffs: map[int][]events.Buff{
					1: {
						&hamster.PlayPowerBuff{Power: 1},
					},
					2: {
						&hamster.PlayPowerBuff{Power: 2},
					},
					3: {
						&hamster.PlayPowerBuff{Power: 3},
					},
					4: {
						&hamster.PlayPowerBuff{Power: 4},
						&hamster.PercentagePowerBuff{Percentage: 0.25},
					},
					5: {
						&hamster.PlayPowerBuff{Power: 5},
						&hamster.PercentagePowerBuff{Percentage: 0.5},
					},
				},
				Prices: map[int]*common.Score{
					1: {
						Mantissa: 250,
					},
					2: {
						Mantissa: 500,
					},
					3: {
						Mantissa: 2000,
					},
					4: {
						Mantissa: 5000,
					},
					5: {
						Mantissa: 10000,
					},
				},
			},
			items.TapCount: {
				Emoji:       "🐹",
				Name:        items.TapCount,
				MaxLevel:    5,
				Description: "Коробка с милыми хомяками.",
				Buffs: map[int][]events.Buff{
					1: {
						&hamster.PlayCountBuff{Count: 10},
						&hamster.PlayPowerBuff{Power: 1},
					},
					2: {
						&hamster.PlayCountBuff{Count: 20},
						&hamster.PlayPowerBuff{Power: 2},
					},
					3: {
						&hamster.PlayCountBuff{Count: 30},
						&hamster.PlayPowerBuff{Power: 3},
					},
					4: {
						&hamster.PlayCountBuff{Count: 40},
						&hamster.PercentagePowerBuff{Percentage: 0.25},
						&hamster.PlayPowerBuff{Power: 4},
					},
					5: {
						&hamster.PlayCountBuff{Count: 50},
						&hamster.PercentagePowerBuff{Percentage: 0.5},
						&hamster.PlayPowerBuff{Power: 5},
					},
				},
				Prices: map[int]*common.Score{
					1: {
						Mantissa: 1000,
					},
					2: {
						Mantissa: 2500,
					},
					3: {
						Mantissa: 5000,
					},
					4: {
						Mantissa: 10000,
					},
					5: {
						Mantissa: 25000,
					},
				},
			},
			items.TapPower: {
				Name:        items.TapPower,
				Emoji:       "🏋️",
				MaxLevel:    5,
				Description: "Тренажер для хомяков, увеличивающий их силу.",
				Buffs: map[int][]events.Buff{
					1: {
						&hamster.PlayPowerBuff{Power: 1},
						&hamster.PercentagePowerBuff{Percentage: 1.0},
					},
					2: {
						&hamster.PlayPowerBuff{Power: 2},
						&hamster.PercentagePowerBuff{Percentage: 1.25},
					},
					3: {
						&hamster.PlayPowerBuff{Power: 3},
						&hamster.PercentagePowerBuff{Percentage: 1.5},
					},
					4: {
						&hamster.PlayPowerBuff{Power: 4},
						&hamster.PercentagePowerBuff{Percentage: 1.75},
					},
					5: {
						&hamster.PlayPowerBuff{Power: 5},
						&hamster.PercentagePowerBuff{Percentage: 2.5},
					},
				},
				Prices: map[int]*common.Score{
					1: {
						Mantissa: 2000,
					},
					2: {
						Mantissa: 5000,
					},
					3: {
						Mantissa: 10000,
					},
					4: {
						Mantissa: 20000,
					},
					5: {
						Mantissa: 50000,
					},
				},
			},
		},
	}
}
