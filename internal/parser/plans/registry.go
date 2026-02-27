package plans

import (
	"fmt"
	"go_parser/internal/domain/plan"
	"sync"
)

type PlanRegistr struct {
	mu    sync.RWMutex
	plans map[string]plan.Plan
}

func NewRegistr() *PlanRegistr {
	return &PlanRegistr{
		plans: make(map[string]plan.Plan),
	}
}

func (r *PlanRegistr) Register(plan plan.Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plan.Name()
	if _, exists := r.plans[name]; exists {
		return fmt.Errorf("plan %s already registered", name)
	}

	r.plans[name] = plan
	return nil
}

func (r *PlanRegistr) Get(name string) (plan.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.plans[name]
	if !exists {
		return nil, fmt.Errorf("plan %s not found", name)
	}
	return plan, nil
}

func (r *PlanRegistr) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plans))
	for name := range r.plans {
		names = append(names, name)
	}
	return names
}
