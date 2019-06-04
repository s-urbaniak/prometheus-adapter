package promql

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

type Enforcer struct {
	entries map[string]*labels.Matcher
}

func NewEnforcer(matchers ...*labels.Matcher) *Enforcer {
	entries := make(map[string]*labels.Matcher)

	for _, matcher := range matchers {
		entries[matcher.Name] = matcher
	}

	return &Enforcer{
		entries: entries,
	}
}

// EnforceNode walks the given node recursively
// and enforces the given label enforcer on it.
//
// Whenever a promql.MatrixSelector or promql.VectorSelector AST node is found,
// their label enforcer are being potentially modified.
//
// If a node label matcher equals the name with one of the given enforcer,
// it is being replaced.
func (ms Enforcer) EnforceNode(node promql.Node) error {
	switch n := node.(type) {
	case *promql.EvalStmt:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case promql.Expressions:
		for _, e := range n {
			if err := ms.EnforceNode(e); err != nil {
				return err
			}
		}

	case *promql.AggregateExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *promql.BinaryExpr:
		if err := ms.EnforceNode(n.LHS); err != nil {
			return err
		}

		if err := ms.EnforceNode(n.RHS); err != nil {
			return err
		}

	case *promql.Call:
		if err := ms.EnforceNode(n.Args); err != nil {
			return err
		}

	case *promql.SubqueryExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *promql.ParenExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *promql.UnaryExpr:
		if err := ms.EnforceNode(n.Expr); err != nil {
			return err
		}

	case *promql.NumberLiteral, *promql.StringLiteral:
	// nothing to do

	case *promql.MatrixSelector:
		// inject labelselector
		n.LabelMatchers = ms.enforceMatchers(n.LabelMatchers)

	case *promql.VectorSelector:
		// inject labelselector
		n.LabelMatchers = ms.enforceMatchers(n.LabelMatchers)

	default:
		panic(fmt.Errorf("promql.Walk: unhandled node type %T", n))
	}

	return nil
}

func (ms Enforcer) enforceMatchers(targets []*labels.Matcher) []*labels.Matcher {
	var res []*labels.Matcher

	for _, target := range targets {
		replacement, ok := ms.entries[target.Name]
		if ok {
			res = append(res, replacement)
		} else {
			res = append(res, target)
		}
	}

	return res
}
