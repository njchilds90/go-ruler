package ruler

// Builder provides a fluent API for assembling rules.
type Builder struct {
	rule Rule
}

// NewRule starts a fluent rule builder.
func NewRule(name string) *Builder { return &Builder{rule: Rule{Name: name, Op: OpAnd}} }

// Description sets rule description.
func (b *Builder) Description(v string) *Builder { b.rule.Description = v; return b }

// Priority sets rule priority.
func (b *Builder) Priority(v Priority) *Builder { b.rule.Priority = v; return b }

// Score sets rule score.
func (b *Builder) Score(v float64) *Builder { b.rule.Score = v; return b }

// Op sets AND/OR condition semantics.
func (b *Builder) Op(v ConditionOp) *Builder { b.rule.Op = v; return b }

// Tag appends a tag.
func (b *Builder) Tag(v string) *Builder { b.rule.Tags = append(b.rule.Tags, v); return b }

// Condition appends a condition.
func (b *Builder) Condition(c Condition) *Builder {
	b.rule.Conditions = append(b.rule.Conditions, c)
	return b
}

// Build finalizes the rule.
func (b *Builder) Build() Rule { return b.rule }
