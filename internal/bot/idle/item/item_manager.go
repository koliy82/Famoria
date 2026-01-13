package item

import (
	"famoria/internal/bot/idle/event"
	"famoria/internal/bot/idle/event/anubis"
	"famoria/internal/bot/idle/event/casino"
	"famoria/internal/bot/idle/event/growkid"
	"famoria/internal/bot/idle/event/hamster"
	"famoria/internal/bot/idle/event/subscribe"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/pkg/common"

	"go.uber.org/zap"
)

type Manager struct {
	Log   *zap.Logger
	Items map[items.ItemId]*Item
}

func (m *Manager) GetItem(name items.ItemId) *Item {
	item := m.Items[name]
	if item == nil {
		m.Log.Sugar().Error("Item not found", name)
	}
	return item
}

type Item struct {
	ItemId      items.ItemId `bson:"name"`
	Emoji       string
	MaxLevel    int
	Buffs       map[int][]event.Buff
	Description string
	Prices      map[int]*common.Score
}

func New(log *zap.Logger) *Manager {
	return &Manager{
		Log: log,
		Items: map[items.ItemId]*Item{
			// Donate items
			items.Subscribe: {
				Emoji:       "💎",
				ItemId:      items.Subscribe,
				MaxLevel:    0,
				Description: "Древний артефакт, испускающий мощную магическую ауру. Этот кристалл дарует владельцу невероятное везение и усиливает все его способности. Легенда гласит, что тот, кто овладеет кристаллом, сможет изменить судьбу своего рода.",
				Buffs: map[int][]event.Buff{
					0: {
						&hamster.PercentagePowerBuff{Percentage: 1.0},
						&casino.PercentagePowerBuff{Percentage: 1.0},
						&growkid.PercentagePowerBuff{Percentage: 1.0},
						&casino.LuckBuff{Luck: 15},
						&subscribe.SaleBuff{Percentage: 0.2},
						&anubis.AccessBuff{},
					},
				},
				Prices: map[int]*common.Score{
					0: {Mantissa: 1000000, Exponent: 100},
				},
			},
			// Hamster items
			items.MegaTap: {
				Emoji:       "💪",
				ItemId:      items.MegaTap,
				MaxLevel:    5,
				Description: "Священная перчатка, усиливающая силу вашего тапа.",
				Buffs: map[int][]event.Buff{
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
				ItemId:      items.TapCount,
				MaxLevel:    5,
				Description: "Коробка с милыми хомяками.",
				Buffs: map[int][]event.Buff{
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
				ItemId:      items.TapPower,
				Emoji:       "🏋️",
				MaxLevel:    5,
				Description: "Тренажер для хомяков, увеличивающий их силу.",
				Buffs: map[int][]event.Buff{
					1: {
						&hamster.PlayPowerBuff{Power: 2},
						&hamster.PercentagePowerBuff{Percentage: 1.0},
					},
					2: {
						&hamster.PlayPowerBuff{Power: 5},
						&hamster.PercentagePowerBuff{Percentage: 1.25},
					},
					3: {
						&hamster.PlayPowerBuff{Power: 7},
						&hamster.PercentagePowerBuff{Percentage: 1.5},
					},
					4: {
						&hamster.PlayPowerBuff{Power: 10},
						&hamster.PercentagePowerBuff{Percentage: 1.75},
						&hamster.PlayCountBuff{Count: 3},
					},
					5: {
						&hamster.PlayPowerBuff{Power: 15},
						&hamster.PercentagePowerBuff{Percentage: 2.5},
						&hamster.PlayCountBuff{Count: 5},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 2_000},
					2: {Mantissa: 5_000},
					3: {Mantissa: 10_000},
					4: {Mantissa: 20_000},
					5: {Mantissa: 50_000},
				},
			},
			items.HamsterWheel: {
				Emoji:       "🏃‍♂️",
				ItemId:      items.HamsterWheel,
				MaxLevel:    5,
				Description: "Колесо хомяка, которое увеличивает скорость и силу их тренировок.",
				Buffs: map[int][]event.Buff{
					1: {
						&hamster.PlayPowerBuff{Power: 2},
					},
					2: {
						&hamster.PlayPowerBuff{Power: 10},
					},
					3: {
						&hamster.PlayPowerBuff{Power: 12},
						&hamster.PlayCountBuff{Count: 1},
					},
					4: {
						&hamster.PlayPowerBuff{Power: 15},
						&hamster.PlayCountBuff{Count: 2},
						&hamster.PercentagePowerBuff{Percentage: 0.2},
					},
					5: {
						&hamster.PlayPowerBuff{Power: 20},
						&hamster.PlayCountBuff{Count: 10},
						&hamster.PercentagePowerBuff{Percentage: 0.4},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 500},
					2: {Mantissa: 1000},
					3: {Mantissa: 2000},
					4: {Mantissa: 5000},
					5: {Mantissa: 10_000},
				},
			},
			items.HamsterCape: {
				Emoji:       "🦸‍♂️",
				ItemId:      items.HamsterCape,
				MaxLevel:    5,
				Description: "Плащ супергероя для хомяков, который придаёт невероятную силу каждому действию.",
				Buffs: map[int][]event.Buff{
					1: {
						&hamster.PercentagePowerBuff{Percentage: 1.0},
					},
					2: {
						&hamster.PercentagePowerBuff{Percentage: 1.5},
						&hamster.PlayPowerBuff{Power: 10},
					},
					3: {
						&hamster.PercentagePowerBuff{Percentage: 2.0},
						&hamster.PlayPowerBuff{Power: 20},
					},
					4: {
						&hamster.PercentagePowerBuff{Percentage: 2.5},
						&hamster.PlayPowerBuff{Power: 25},
					},
					5: {
						&hamster.PercentagePowerBuff{Percentage: 3.0},
						&hamster.PlayPowerBuff{Power: 50},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 50_000},
					2: {Mantissa: 100_500},
					3: {Mantissa: 150_000},
					4: {Mantissa: 200_000},
					5: {Mantissa: 500_000},
				},
			},

			// Casino items
			items.GoldenDice: {
				Emoji:       "🎲",
				ItemId:      items.GoldenDice,
				MaxLevel:    5,
				Description: "Эти золотые кости, выкованные богами удачи, увеличивают твой выигрыш на каждом броске.",
				Buffs: map[int][]event.Buff{
					1: {
						&casino.PlayPowerBuff{Power: 1000},
					},
					2: {
						&casino.PlayPowerBuff{Power: 2000},
						&casino.PercentagePowerBuff{Percentage: 0.1},
					},
					3: {
						&casino.PlayPowerBuff{Power: 3000},
						&casino.PercentagePowerBuff{Percentage: 0.25},
					},
					4: {
						&casino.PlayPowerBuff{Power: 5000},
						&casino.PercentagePowerBuff{Percentage: 0.5},
					},
					5: {
						&casino.PlayPowerBuff{Power: 10000},
						&casino.PercentagePowerBuff{Percentage: 1.0},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 2_500},
					2: {Mantissa: 5_000},
					3: {Mantissa: 10_000},
					4: {Mantissa: 25_000},
					5: {Mantissa: 100_000},
				},
			},
			items.InfiniteSpins: {
				Emoji:       "🔄",
				ItemId:      items.InfiniteSpins,
				MaxLevel:    5,
				Description: "Эти магические барабаны могут вращаться вечно, увеличивая количество твоих попыток.",
				Buffs: map[int][]event.Buff{
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
				ItemId:      items.LuckyClover,
				MaxLevel:    5,
				Description: "Легендарный клевер находит счастливчика среди всех и делает его ещё удачливее!",
				Buffs: map[int][]event.Buff{
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
					1: {Mantissa: 100_000},
					2: {Mantissa: 250_000},
					3: {Mantissa: 500_000},
					4: {Mantissa: 1_000_000},
					5: {Mantissa: 2_500_000},
				},
			},
			items.LuckyCharm: {
				Emoji:       "🧲",
				ItemId:      items.LuckyCharm,
				MaxLevel:    5,
				Description: "Амулет удачи, притягивающий счастливые моменты и увеличивающий шанс на выигрыш.",
				Buffs: map[int][]event.Buff{
					1: {
						&casino.LuckBuff{Luck: 2},
					},
					2: {
						&casino.LuckBuff{Luck: 3},
						&casino.PercentagePowerBuff{Percentage: 0.1},
					},
					3: {
						&casino.LuckBuff{Luck: 5},
						&casino.PercentagePowerBuff{Percentage: 0.1},
						&casino.PlayPowerBuff{Power: 50},
					},
					4: {
						&casino.LuckBuff{Luck: 7},
						&casino.PercentagePowerBuff{Percentage: 0.2},
						&casino.PlayPowerBuff{Power: 50},
					},
					5: {
						&casino.LuckBuff{Luck: 10},
						&casino.PercentagePowerBuff{Percentage: 0.3},
						&casino.PlayPowerBuff{Power: 100},
						&casino.PlayCountBuff{Count: 1},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 10_000},
					2: {Mantissa: 25_000},
					3: {Mantissa: 50_000},
					4: {Mantissa: 200_000},
					5: {Mantissa: 500_000},
				},
			},
			items.FortuneTalisman: {
				Emoji:       "🧿",
				ItemId:      items.FortuneTalisman,
				MaxLevel:    5,
				Description: "Талисман удачи, который притягивает богатство и усиливает все действия в казино.",
				Buffs: map[int][]event.Buff{
					1: {
						&casino.LuckBuff{Luck: 10},
					},
					2: {
						&casino.LuckBuff{Luck: 15},
						&casino.PlayPowerBuff{Power: 100},
					},
					3: {
						&casino.LuckBuff{Luck: 20},
						&casino.PlayPowerBuff{Power: 200},
						&casino.PercentagePowerBuff{Percentage: 0.2},
					},
					4: {
						&casino.LuckBuff{Luck: 25},
						&casino.PlayPowerBuff{Power: 300},
						&casino.PercentagePowerBuff{Percentage: 0.5},
					},
					5: {
						&casino.LuckBuff{Luck: 30},
						&casino.PlayPowerBuff{Power: 500},
						&casino.PercentagePowerBuff{Percentage: 1.0},
						&casino.PlayCountBuff{Count: 1},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 1_000_000},
					2: {Mantissa: 2_000_000},
					3: {Mantissa: 5_000_000},
					4: {Mantissa: 10_000_000},
					5: {Mantissa: 25_000_000},
				},
			},

			// Grow items
			items.MagicSpoon: {
				Emoji:       "🥄",
				ItemId:      items.MagicSpoon,
				MaxLevel:    5,
				Description: "Эта ложка, выкованная из звёздного света, увеличивает эффект каждого кормления.",
				Buffs: map[int][]event.Buff{
					1: {
						&growkid.PlayPowerBuff{Power: 100},
					},
					2: {
						&growkid.PlayPowerBuff{Power: 250},
					},
					3: {
						&growkid.PlayPowerBuff{Power: 1000},
					},
					4: {
						&growkid.PlayPowerBuff{Power: 2500},
						&growkid.PercentagePowerBuff{Percentage: 0.1},
					},
					5: {
						&growkid.PlayPowerBuff{Power: 5000},
						&growkid.PercentagePowerBuff{Percentage: 0.25},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 100},
					2: {Mantissa: 2500},
					3: {Mantissa: 5000},
					4: {Mantissa: 10000},
					5: {Mantissa: 50000},
				},
			},
			items.GrowthPotion: {
				Emoji:       "🧪",
				ItemId:      items.GrowthPotion,
				MaxLevel:    5,
				Description: "Эликсир, сваренный древним алхимиком, ускоряет рост ребёнка.",
				Buffs: map[int][]event.Buff{
					1: {
						&growkid.PercentagePowerBuff{Percentage: 0.25},
					},
					2: {
						&growkid.PercentagePowerBuff{Percentage: 0.5},
						&growkid.PlayPowerBuff{Power: 50},
					},
					3: {
						&growkid.PercentagePowerBuff{Percentage: 1.0},
						&growkid.PlayPowerBuff{Power: 100},
					},
					4: {
						&growkid.PercentagePowerBuff{Percentage: 1.5},
						&growkid.PlayPowerBuff{Power: 150},
					},
					5: {
						&growkid.PercentagePowerBuff{Percentage: 2.0},
						&growkid.PlayPowerBuff{Power: 250},
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
				ItemId:      items.EndlessMilk,
				MaxLevel:    5,
				Description: "Бутылочка молока, которое никогда не заканчивается, увеличивая количество кормлений.",
				Buffs: map[int][]event.Buff{
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
			items.FertilityRing: {
				Emoji:       "💍",
				ItemId:      items.FertilityRing,
				MaxLevel:    5,
				Description: "Магическое кольцо, которое ускоряет рост ребенка и улучшает его состояние.",
				Buffs: map[int][]event.Buff{
					1: {
						&growkid.PlayPowerBuff{Power: 250},
						&growkid.PercentagePowerBuff{Percentage: 0.2},
					},
					2: {
						&growkid.PlayPowerBuff{Power: 500},
						&growkid.PercentagePowerBuff{Percentage: 0.5},
					},
					3: {
						&growkid.PlayPowerBuff{Power: 1500},
						&growkid.PercentagePowerBuff{Percentage: 0.75},
					},
					4: {
						&growkid.PlayPowerBuff{Power: 3000},
						&growkid.PercentagePowerBuff{Percentage: 1.0},
					},
					5: {
						&growkid.PlayPowerBuff{Power: 5000},
						&growkid.PercentagePowerBuff{Percentage: 1.0},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 2000},
					2: {Mantissa: 5000},
					3: {Mantissa: 10_000},
					4: {Mantissa: 20_000},
					5: {Mantissa: 50_000},
				},
			},
			items.MagicSunflower: {
				Emoji:       "🌻",
				ItemId:      items.MagicSunflower,
				MaxLevel:    5,
				Description: "Волшебный подсолнух, излучающий свет, который ускоряет рост ребёнка и увеличивает эффективность тренировок.",
				Buffs: map[int][]event.Buff{
					1: {
						&growkid.PlayPowerBuff{Power: 1500},
						&growkid.PercentagePowerBuff{Percentage: 0.5},
					},
					2: {
						&growkid.PlayPowerBuff{Power: 5000},
						&growkid.PercentagePowerBuff{Percentage: 1.0},
					},
					3: {
						&growkid.PlayPowerBuff{Power: 10000},
						&growkid.PercentagePowerBuff{Percentage: 2.0},
					},
					4: {
						&growkid.PlayPowerBuff{Power: 20000},
						&growkid.PercentagePowerBuff{Percentage: 2.5},
					},
					5: {
						&growkid.PlayPowerBuff{Power: 50000},
						&growkid.PercentagePowerBuff{Percentage: 3.0},
						&growkid.PlayCountBuff{Count: 1},
					},
				},
				Prices: map[int]*common.Score{
					1: {Mantissa: 300_000},
					2: {Mantissa: 1_200_000},
					3: {Mantissa: 3_500_000},
					4: {Mantissa: 5_000_000},
					5: {Mantissa: 10_000_000},
				},
			},
		},
	}
}
