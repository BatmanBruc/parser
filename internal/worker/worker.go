package worker

import (
	"encoding/json"
	"go_parser/internal/domain/plan"
	"go_parser/internal/domain/queue"
	"go_parser/internal/domain/task"
	"go_parser/internal/utils"
)

type PlanRegister interface {
	Get(name string) (plan.Plan, error)
}

type Handler interface {
	HandleResult(result *plan.PlanResult, foundURLs []plan.FoundURL, err error) error
}

type WorkerPool struct {
	Msg     chan queue.WrapperMessage
	quit    chan struct{}
	planReg PlanRegister
	count   int
	h       Handler
}

func NewWorkerPool(count int, planReg PlanRegister, h Handler) *WorkerPool {
	return &WorkerPool{
		Msg:     make(chan queue.WrapperMessage),
		quit:    make(chan struct{}),
		planReg: planReg,
		count:   count,
		h:       h,
	}
}

func (w WorkerPool) Start() {
	for i := 0; i < w.count; i++ {
		go func() {
			for {
				select {
				case msg := <-w.Msg:
					w.proccesTask(msg)
				case <-w.quit:
					return
				}
			}
		}()
	}
}

func (w *WorkerPool) proccesTask(msg queue.WrapperMessage) {
	var task *task.Task

	if err := json.Unmarshal(msg.GetBody(), &task); err != nil {
		utils.Logger.Printf("Ошибка разбора сообщения: %v\n", err)
		res := &plan.PlanResult{
			URL:      task.URL,
			PlanName: task.Plan,
			Error:    err.Error(),
		}
		var urls []plan.FoundURL
		w.h.HandleResult(res, urls, err)
		msg.Reject()
	}

	pln, err := w.planReg.Get(task.Plan)

	if err != nil {
		utils.Logger.Printf("Ошибка разбора сообщения: %v\n", err)
		res := &plan.PlanResult{
			URL:      task.URL,
			PlanName: task.Plan,
			Error:    err.Error(),
		}
		var urls []plan.FoundURL
		w.h.HandleResult(res, urls, err)
		msg.Reject()
	}

	res, urls, err := pln.Execute(task)

	err = w.h.HandleResult(res, urls, err)
	if err != nil {
		utils.Logger.Printf("Обработка результата выполнена успешно")
	}
}

func (w *WorkerPool) Stop() {
	close(w.quit)
}
