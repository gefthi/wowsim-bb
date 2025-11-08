package apl

import "time"

// EvaluationContext is provided by the simulator when evaluating conditions.
// The implementation will be added when the engine integrates the APL system.
type EvaluationContext interface {
	BuffActive(name string) bool
	BuffRemaining(name string) time.Duration
	BuffCharges(name string) int
	DebuffActive(name string) bool
	DebuffRemaining(name string) time.Duration
	ResourcePercent(resource string) float64
	CooldownReady(name string) bool
	CooldownRemaining(name string) time.Duration
}

// Condition evaluates to true/false for a given context.
type Condition interface {
	Eval(ctx EvaluationContext) bool
}

// Always true/false conditions.
type trueCondition struct{}

func (trueCondition) Eval(EvaluationContext) bool { return true }

type falseCondition struct{}

func (falseCondition) Eval(EvaluationContext) bool { return false }

// anyCondition is logical OR.
type anyCondition struct {
	children []Condition
}

func (c anyCondition) Eval(ctx EvaluationContext) bool {
	for _, child := range c.children {
		if child.Eval(ctx) {
			return true
		}
	}
	return false
}

// allCondition is logical AND.
type allCondition struct {
	children []Condition
}

func (c allCondition) Eval(ctx EvaluationContext) bool {
	for _, child := range c.children {
		if !child.Eval(ctx) {
			return false
		}
	}
	return true
}

// notCondition negates a child.
type notCondition struct {
	child Condition
}

func (c notCondition) Eval(ctx EvaluationContext) bool {
	if c.child == nil {
		return true
	}
	return !c.child.Eval(ctx)
}

// debuffActiveCondition checks if a debuff/buff on target is active.
type debuffActiveCondition struct {
	name         string
	minRemaining *time.Duration
	maxRemaining *time.Duration
}

func (c debuffActiveCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	if !ctx.DebuffActive(c.name) {
		return false
	}
	remaining := ctx.DebuffRemaining(c.name)
	if c.minRemaining != nil && remaining < *c.minRemaining {
		return false
	}
	if c.maxRemaining != nil && remaining > *c.maxRemaining {
		return false
	}
	return true
}

// dotRemainingCondition compares DoT remaining duration.
type dotRemainingCondition struct {
	spell string
	lt    *time.Duration
	gt    *time.Duration
	lte   *time.Duration
	gte   *time.Duration
}

func (c dotRemainingCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	remaining := ctx.DebuffRemaining(c.spell)
	if c.lt != nil && !(remaining < *c.lt) {
		return false
	}
	if c.lte != nil && !(remaining <= *c.lte) {
		return false
	}
	if c.gt != nil && !(remaining > *c.gt) {
		return false
	}
	if c.gte != nil && !(remaining >= *c.gte) {
		return false
	}
	return true
}

// resourcePercentCondition compares resource levels (mana, health, etc).
type resourcePercentCondition struct {
	resource string
	lt       *float64
	lte      *float64
	gt       *float64
	gte      *float64
}

func (c resourcePercentCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	percent := ctx.ResourcePercent(c.resource)
	if c.lt != nil && !(percent < *c.lt) {
		return false
	}
	if c.lte != nil && !(percent <= *c.lte) {
		return false
	}
	if c.gt != nil && !(percent > *c.gt) {
		return false
	}
	if c.gte != nil && !(percent >= *c.gte) {
		return false
	}
	return true
}

// cooldownReadyCondition checks if a spell/item is off cooldown.
type cooldownReadyCondition struct {
	name string
}

func (c cooldownReadyCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	return ctx.CooldownReady(c.name)
}

type cooldownRemainingCondition struct {
	name string
	lt   *time.Duration
	lte  *time.Duration
	gt   *time.Duration
	gte  *time.Duration
}

func (c cooldownRemainingCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	remaining := ctx.CooldownRemaining(c.name)
	if c.lt != nil && !(remaining < *c.lt) {
		return false
	}
	if c.lte != nil && !(remaining <= *c.lte) {
		return false
	}
	if c.gt != nil && !(remaining > *c.gt) {
		return false
	}
	if c.gte != nil && !(remaining >= *c.gte) {
		return false
	}
	return true
}

type buffActiveCondition struct {
	name         string
	minRemaining *time.Duration
	maxRemaining *time.Duration
}

func (c buffActiveCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	if !ctx.BuffActive(c.name) {
		return false
	}
	remaining := ctx.BuffRemaining(c.name)
	if c.minRemaining != nil && remaining < *c.minRemaining {
		return false
	}
	if c.maxRemaining != nil && remaining > *c.maxRemaining {
		return false
	}
	return true
}

type chargesCondition struct {
	buff string
	lt   *int
	lte  *int
	gt   *int
	gte  *int
}

func (c chargesCondition) Eval(ctx EvaluationContext) bool {
	if ctx == nil {
		return false
	}
	charges := ctx.BuffCharges(c.buff)
	if c.lt != nil && !(charges < *c.lt) {
		return false
	}
	if c.lte != nil && !(charges <= *c.lte) {
		return false
	}
	if c.gt != nil && !(charges > *c.gt) {
		return false
	}
	if c.gte != nil && !(charges >= *c.gte) {
		return false
	}
	return true
}
