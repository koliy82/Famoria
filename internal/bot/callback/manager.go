package callback

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

type CallbacksManager struct {
	mu        sync.Mutex
	Callbacks map[string]Callback
	log       *zap.Logger
}

func New(log *zap.Logger) *CallbacksManager {
	return &CallbacksManager{
		Callbacks: make(map[string]Callback),
		log:       log,
	}
}

func (cm *CallbacksManager) StaticCallback(data string, callback func(query telego.CallbackQuery)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.Callbacks[data]; exists {
		cm.log.Sugar().Fatalf("Callback with data %s already exists", data)
	}
	cm.Callbacks[data] = Callback{
		Data:     data,
		Type:     Static,
		OwnerIDs: []int64{},
		Callback: callback,
	}
}

type DynamicOpts struct {
	Label      string
	CtxType    ContextType
	OwnerIDs   []int64
	Time       time.Duration
	AnswerText string
	Callback   func(query telego.CallbackQuery)
}

func (cm *CallbacksManager) DynamicCallback(opts DynamicOpts) Callback {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	uid, _ := uuid.NewUUID()
	data := uid.String()
	if _, exists := cm.Callbacks[data]; exists {
		cm.log.Sugar().Fatalf("Callback with data %s already exists", data)
	}

	newCallback := Callback{
		Data:       data,
		Type:       opts.CtxType,
		OwnerIDs:   opts.OwnerIDs,
		Label:      opts.Label,
		AnswerText: opts.AnswerText,
		Callback:   opts.Callback,
	}

	cm.Callbacks[data] = newCallback

	go cm.cleanupCallback(data, opts.Time)

	return newCallback
}

func (cm *CallbacksManager) cleanupCallback(data string, duration time.Duration) {
	time.Sleep(duration)
	cm.RemoveCallback(data)
}

func (cm *CallbacksManager) RemoveCallback(data string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.Callbacks, data)
}

func contains(slice []int64, item int64) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (cm *CallbacksManager) HandleCallback(ctx *th.Context, query telego.CallbackQuery) {
	cm.mu.Lock()
	callback, exists := cm.Callbacks[query.Data]
	cm.mu.Unlock()

	params := &telego.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "У кнопки истёк срок действия.",
	}

	if !exists {
		err := ctx.Bot().AnswerCallbackQuery(
			context.Background(),
			params.WithText("У кнопки истёк срок действия."),
		)
		if err != nil {
			cm.log.Sugar().Error(err)
		}
		return
	}

	if len(callback.OwnerIDs) > 0 && !contains(callback.OwnerIDs, query.From.ID) {
		err := ctx.Bot().AnswerCallbackQuery(
			context.Background(),
			&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Кнопка запривачена, ты не можешь её нажать.",
			},
		)
		if err != nil {
			cm.log.Sugar().Error(err)
		}
		return
	}

	callback.Callback(query)

	if callback.Type != Static {
		err := ctx.Bot().AnswerCallbackQuery(
			context.Background(),
			params.WithText(callback.AnswerText),
		)
		if err != nil {
			cm.log.Sugar().Error(err)
		}
	}

	switch callback.Type {
	case Static, Temporary:
	case OneClick:
		cm.RemoveCallback(query.Data)
	case ChooseOne:
		for data, cb := range cm.Callbacks {
			if cb.Type == ChooseOne && contains(cb.OwnerIDs, query.From.ID) {
				cm.RemoveCallback(data)
			}
		}
	}
}
