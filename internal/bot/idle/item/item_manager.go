package item

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/events/casino"
	"famoria/internal/bot/idle/events/growkid"
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
			// Hamster items
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
					1: {Mantissa: 250},
					2: {Mantissa: 500},
					3: {Mantissa: 2000},
					4: {Mantissa: 5000},
					5: {Mantissa: 10000},
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
					1: {Mantissa: 1000},
					2: {Mantissa: 2500},
					3: {Mantissa: 5000},
					4: {Mantissa: 10000},
					5: {Mantissa: 25000},
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
					1: {Mantissa: 2000},
					2: {Mantissa: 5000},
					3: {Mantissa: 10000},
					4: {Mantissa: 20000},
					5: {Mantissa: 50000},
				},
			},

			// Casino items
			items.GoldenDice: {
				Emoji:       "🎲",
				Name:        items.GoldenDice,
				MaxLevel:    5,
				Description: "Эти золотые кости, выкованные богами удачи, увеличивают твой выигрыш на каждом броске.",
				Buffs: map[int][]events.Buff{
					1: {
						&casino.PlayPowerBuff{Power: 250},
					},
					2: {
						&casino.PlayPowerBuff{Power: 500},
						&casino.PercentagePowerBuff{Percentage: 0.05},
					},
					3: {
						&casino.PlayPowerBuff{Power: 750},
						&casino.PercentagePowerBuff{Percentage: 0.1},
					},
					4: {
						&casino.PlayPowerBuff{Power: 1000},
						&casino.PercentagePowerBuff{Percentage: 0.25},
					},
					5: {
						&casino.PlayPowerBuff{Power: 1500},
						&casino.PercentagePowerBuff{Percentage: 0.3},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 2000},
					2: {Mantissa: 5000},
					3: {Mantissa: 10000},
					4: {Mantissa: 20000},
					5: {Mantissa: 50000},
				},
			},
			items.InfiniteSpins: {
				Emoji:       "🔄",
				Name:        items.InfiniteSpins,
				MaxLevel:    5,
				Description: "Эти магические барабаны могут вращаться вечно, увеличивая количество твоих попыток.",
				Buffs: map[int][]events.Buff{
					1: {
						&casino.PlayCountBuff{Count: 1},
					},
					2: {
						&casino.PlayCountBuff{Count: 2},
						&casino.PlayPowerBuff{Power: 10},
					},
					3: {
						&casino.PlayCountBuff{Count: 3},
						&casino.PlayPowerBuff{Power: 50},
					},
					4: {
						&casino.PlayCountBuff{Count: 4},
						&casino.PlayPowerBuff{Power: 100},
						&casino.PercentagePowerBuff{Percentage: 0.1},
					},
					5: {
						&casino.PlayCountBuff{Count: 5},
						&casino.PlayPowerBuff{Power: 300},
						&casino.PercentagePowerBuff{Percentage: 0.25},
						&casino.LuckBuff{Luck: 5},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 50_000},
					2: {Mantissa: 100_000},
					3: {Mantissa: 250_000},
					4: {Mantissa: 500_000},
					5: {Mantissa: 1_000_000},
				},
			},
			items.LuckyClover: {
				Emoji:       "🍀",
				Name:        items.LuckyClover,
				MaxLevel:    5,
				Description: "Легендарный клевер находит счастливчика среди всех и делает его ещё удачливее!",
				Buffs: map[int][]events.Buff{
					1: {
						&casino.LuckBuff{Luck: 10},
					},
					2: {
						&casino.LuckBuff{Luck: 15},
						&casino.PlayPowerBuff{Power: 10},
					},
					3: {
						&casino.LuckBuff{Luck: 20},
						&casino.PlayPowerBuff{Power: 50},
					},
					4: {
						&casino.LuckBuff{Luck: 25},
						&casino.PlayPowerBuff{Power: 100},
						&casino.PercentagePowerBuff{Percentage: 0.25},
					},
					5: {
						&casino.LuckBuff{Luck: 30},
						&casino.PlayPowerBuff{Power: 300},
						&casino.PercentagePowerBuff{Percentage: 0.5},
						&casino.PlayCountBuff{Count: 1},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 50_000},
					2: {Mantissa: 100_000},
					3: {Mantissa: 250_000},
					4: {Mantissa: 500_000},
					5: {Mantissa: 1_000_000},
				},
			},

			// Grow items
			items.MagicSpoon: {
				Emoji:       "🥄",
				Name:        items.MagicSpoon,
				MaxLevel:    5,
				Description: "Эта ложка, выкованная из звёздного света, увеличивает эффект каждого кормления.",
				Buffs: map[int][]events.Buff{
					1: {
						&growkid.PlayPowerBuff{Power: 50},
					},
					2: {
						&growkid.PlayPowerBuff{Power: 100},
					},
					3: {
						&growkid.PlayPowerBuff{Power: 150},
					},
					4: {
						&growkid.PlayPowerBuff{Power: 200},
						&growkid.PercentagePowerBuff{Percentage: 0.05},
					},
					5: {
						&growkid.PlayPowerBuff{Power: 250},
						&growkid.PercentagePowerBuff{Percentage: 0.1},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 100},
					2: {Mantissa: 500},
					3: {Mantissa: 1000},
					4: {Mantissa: 2500},
					5: {Mantissa: 5000},
				},
			},
			items.GrowthPotion: {
				Emoji:       "🧪",
				Name:        items.GrowthPotion,
				MaxLevel:    5,
				Description: "Эликсир, сваренный древним алхимиком, ускоряет рост ребёнка.",
				Buffs: map[int][]events.Buff{
					1: {
						&growkid.PercentagePowerBuff{Percentage: 0.25},
					},
					2: {
						&growkid.PercentagePowerBuff{Percentage: 0.35},
						&growkid.PlayPowerBuff{Power: 50},
					},
					3: {
						&growkid.PercentagePowerBuff{Percentage: 0.5},
						&growkid.PlayPowerBuff{Power: 100},
					},
					4: {
						&growkid.PercentagePowerBuff{Percentage: 0.75},
						&growkid.PlayPowerBuff{Power: 150},
					},
					5: {
						&growkid.PercentagePowerBuff{Percentage: 1.0},
						&growkid.PlayPowerBuff{Power: 200},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 1000},
					2: {Mantissa: 2500},
					3: {Mantissa: 5000},
					4: {Mantissa: 10000},
					5: {Mantissa: 15000},
				},
			},
			items.EndlessMilk: {
				Emoji:       "🍼",
				Name:        items.EndlessMilk,
				MaxLevel:    5,
				Description: "Бутылочка молока, которое никогда не заканчивается, увеличивая количество кормлений.",
				Buffs: map[int][]events.Buff{
					1: {
						&growkid.PlayCountBuff{Count: 1},
					},
					2: {
						&growkid.PlayCountBuff{Count: 2},
					},
					3: {
						&growkid.PlayCountBuff{Count: 3},
						&growkid.PercentagePowerBuff{Percentage: 0.1},
					},
					4: {
						&growkid.PlayCountBuff{Count: 4},
						&growkid.PercentagePowerBuff{Percentage: 0.1},
						&growkid.PlayPowerBuff{Power: 50},
					},
					5: {
						&growkid.PlayCountBuff{Count: 5},
						&growkid.PercentagePowerBuff{Percentage: 0.2},
						&growkid.PlayPowerBuff{Power: 100},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 5_000},
					2: {Mantissa: 15_000},
					3: {Mantissa: 35_000},
					4: {Mantissa: 100_000},
					5: {Mantissa: 500_000},
				},
			},
		},
	}
}
