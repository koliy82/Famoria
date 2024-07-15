package callback

import (
	"fmt"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"log"
	"sync"
	"time"
)

type CallbacksManager struct {
	mu        sync.Mutex
	callbacks map[string]Callback
}

func New() *CallbacksManager {
	return &CallbacksManager{
		callbacks: make(map[string]Callback),
	}
}

func (cm *CallbacksManager) StaticCallback(data string, callback func(query telego.CallbackQuery)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.callbacks[data]; exists {
		log.Fatalf("Callback with data %s already exists", data)
	}
	cm.callbacks[data] = Callback{
		Data:     data,
		Type:     Static,
		OwnerIDs: []int64{},
		Callback: callback,
	}
}

func (cm *CallbacksManager) DynamicCallback(label string, ctype ContextType, ownerIDs []int64, minutes int32, answerText string, callback func(query telego.CallbackQuery)) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	data := fmt.Sprintf("%d", time.Now().UnixNano())
	if _, exists := cm.callbacks[data]; exists {
		log.Fatalf("Callback with data %s already exists", data)
	}
	cm.callbacks[data] = Callback{
		Data:       data,
		Type:       ctype,
		OwnerIDs:   ownerIDs,
		Label:      label,
		AnswerText: answerText,
		Callback:   callback,
	}

	go cm.cleanupCallback(data, minutes)

	return data
}

func (cm *CallbacksManager) cleanupCallback(data string, minutes int32) {
	time.Sleep(time.Duration(minutes) * time.Minute)
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.callbacks, data)
}

func contains(slice []int64, item int64) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (cm *CallbacksManager) HandleCallback(bot *telego.Bot, query telego.CallbackQuery, log *zap.Logger) {
	log.Sugar().Debug(query)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	callback, exists := cm.callbacks[query.Data]

	if !exists {
		err := bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "У кнопки истёк срок действия.",
			},
		)
		if err != nil {
			log.Sugar().Error(err)
		}
		return
	}

	if len(callback.OwnerIDs) > 0 && !contains(callback.OwnerIDs, query.From.ID) {
		err := bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Кнопка запривачена, ты не можешь её нажать.",
			},
		)
		if err != nil {
			log.Sugar().Error(err)
		}
		return
	}

	callback.Callback(query)

	if callback.AnswerText != "" {
		err := bot.AnswerCallbackQuery(
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            callback.AnswerText,
			},
		)
		if err != nil {
			log.Sugar().Error(err)
		}
		return
	}

	switch callback.Type {
	case Static, Temporary:
	case OneClick:
		delete(cm.callbacks, query.Data)
	case ChooseOne:
		for data, cb := range cm.callbacks {
			if cb.Type == ChooseOne && contains(cb.OwnerIDs, query.From.ID) {
				delete(cm.callbacks, data)
			}
		}
	}

	err := bot.AnswerCallbackQuery(
		&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		},
	)

	if err != nil {
		log.Sugar().Error(err)
	}
}
