package promql

import (
	"fmt"
	"testing"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

type checkFunc func(expression string, err error) error

func checks(cs ...checkFunc) checkFunc {
	return func(expression string, err error) error {
		for _, c := range cs {
			if e := c(expression, err); e != nil {
				return e
			}
		}
		return nil
	}
}

func hasError(want error) checkFunc {
	return func(_ string, got error) error {
		wantError, gotError := "<nil>", "<nil>"

		if want != nil {
			wantError = fmt.Sprintf("%q", want.Error())
		}

		if got != nil {
			gotError = fmt.Sprintf("%q", got.Error())
		}

		if wantError != gotError {
			return fmt.Errorf("want error %v, got %v", wantError, gotError)
		}

		return nil
	}
}

func hasExpression(want string) checkFunc {
	return func(got string, _ error) error {
		if want != got {
			return fmt.Errorf("want expression %v, got %v", want, got)
		}
		return nil
	}
}

var tests = []struct {
	name       string
	expression string
	enforcer   *Enforcer
	check      checkFunc
}{
	{
		name:       "expressions",
		expression: `round(metric1{label="baz",pod="foo",namespace="bar"},3)`,
		enforcer: NewEnforcer(
			&labels.Matcher{
				Name:  "namespace",
				Type:  labels.MatchEqual,
				Value: "NS",
			},
			&labels.Matcher{
				Name:  "pod",
				Type:  labels.MatchEqual,
				Value: "POD",
			},
		),
		check: checks(
			hasError(nil),
			hasExpression(`round(metric1{label="baz",namespace="NS",pod="POD"}, 3)`),
		),
	},

	{
		name:       "aggregate",
		expression: `sum by (pod) (metric1{label="baz",pod="foo",namespace="bar"})`,
		enforcer: NewEnforcer(
			&labels.Matcher{
				Name:  "namespace",
				Type:  labels.MatchEqual,
				Value: "NS",
			},
			&labels.Matcher{
				Name:  "pod",
				Type:  labels.MatchEqual,
				Value: "POD",
			},
		),
		check: checks(
			hasError(nil),
			hasExpression(`sum by(pod) (metric1{label="baz",namespace="NS",pod="POD"})`),
		),
	},

	{
		name:       "binary expression",
		expression: `metric1{pod="baz"} + sum by (pod)(metric2{label="baz",pod="foo",namespace="bar"})`,
		enforcer: NewEnforcer(
			&labels.Matcher{
				Name:  "namespace",
				Type:  labels.MatchEqual,
				Value: "NS",
			},
			&labels.Matcher{
				Name:  "pod",
				Type:  labels.MatchEqual,
				Value: "POD",
			},
		),
		check: checks(
			hasError(nil),
			hasExpression(`metric1{pod="POD"} + sum by(pod) (metric2{label="baz",namespace="NS",pod="POD"})`),
		),
	},

	{
		name:       "binary expression with vector matching",
		expression: `metric1{pod="baz"} + on(pod,namespace) sum by (pod) (metric2{label="baz",pod="foo",namespace="bar"})`,
		enforcer: NewEnforcer(
			&labels.Matcher{
				Name:  "namespace",
				Type:  labels.MatchEqual,
				Value: "NS",
			},
			&labels.Matcher{
				Name:  "pod",
				Type:  labels.MatchEqual,
				Value: "POD",
			},
		),
		check: checks(
			hasError(nil),
			hasExpression(`metric1{pod="POD"} + on(pod, namespace) sum by(pod) (metric2{label="baz",namespace="NS",pod="POD"})`),
		),
	},
}

func TestEnforceNode(t *testing.T) {
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e, err := promql.ParseExpr(tc.expression)
			if err != nil {
				t.Fatal(err)
			}

			err = tc.enforcer.EnforceNode(e)
			if err := tc.check(e.String(), err); err != nil {
				t.Error(err)
			}
		})
	}
}
