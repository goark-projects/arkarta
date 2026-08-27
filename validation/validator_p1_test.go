package validation_test

import (
	"context"
	"reflect"
	"testing"

	"goark.dev/arkarta/validation"
)

func TestDefaultValidatorValidatesGroupsAndMessages(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name" arkarta:"required" arkarta-groups:"create"`
		Code string `json:"code" arkarta:"required"`
	}
	validator := validation.NewValidator(validation.WithMessageResolver(validation.MessageResolverFunc(func(_ context.Context, message validation.MessageContext) (string, bool) {
		if message.Code() == "required" {
			return message.Path() + " 必填", true
		}
		return "", false
	})))

	createResult, err := validator.ValidateGroups(context.Background(), payload{}, "create")
	if err != nil {
		t.Fatalf("ValidateGroups create failed: %v", err)
	}
	createViolations := createResult.Violations()
	if len(createViolations) != 1 || createViolations[0].Path() != "name" || createViolations[0].Message() != "name 必填" {
		t.Fatalf("create violations = %#v, want only custom name required", createViolations)
	}

	defaultResult, err := validator.Validate(context.Background(), payload{})
	if err != nil {
		t.Fatalf("Validate default failed: %v", err)
	}
	if got := p1ViolationPaths(defaultResult); !reflect.DeepEqual(got, []string{"code:required"}) {
		t.Fatalf("default violations = %#v, want code required", got)
	}
}

func TestDefaultValidatorBuiltInP1Constraints(t *testing.T) {
	t.Parallel()

	type payload struct {
		Pattern    string `json:"pattern" arkarta:"regexp=^[a-z]+-[a-z]+$"`
		URL        string `json:"url" arkarta:"url"`
		UUID       string `json:"uuid" arkarta:"uuid"`
		GT         int    `json:"gt" arkarta:"gt=10"`
		GTE        int    `json:"gte" arkarta:"gte=10"`
		LT         int    `json:"lt" arkarta:"lt=10"`
		LTE        int    `json:"lte" arkarta:"lte=10"`
		Contains   string `json:"contains" arkarta:"contains=ark"`
		StartsWith string `json:"startsWith" arkarta:"startswith=go"`
		EndsWith   string `json:"endsWith" arkarta:"endswith=dev"`
	}

	result, err := validation.Validate(context.Background(), payload{
		Pattern:    "arkarta",
		URL:        "://bad",
		UUID:       "bad",
		GT:         10,
		GTE:        9,
		LT:         10,
		LTE:        11,
		Contains:   "go",
		StartsWith: "arkarta",
		EndsWith:   "arkarta",
	})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	want := []string{
		"pattern:regexp",
		"url:url",
		"uuid:uuid",
		"gt:gt",
		"gte:gte",
		"lt:lt",
		"lte:lte",
		"contains:contains",
		"startsWith:startswith",
		"endsWith:endswith",
	}
	if got := p1ViolationPaths(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %#v, want %#v", got, want)
	}
}

func TestDefaultValidatorObjectConstraint(t *testing.T) {
	t.Parallel()

	type period struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}
	validator := validation.NewValidator(validation.WithObjectConstraint(period{}, validation.ObjectConstraintFunc{
		ConstraintName: "periodOrder",
		Fn: func(_ context.Context, object validation.ObjectContext) ([]validation.Violation, error) {
			value := object.Interface().(period)
			if value.End < value.Start {
				return []validation.Violation{
					validation.NewViolation(object.FieldPath("end"), "periodOrder", "结束值不能小于开始值", value.End),
				}, nil
			}
			return nil, nil
		},
	}))

	result, err := validator.Validate(context.Background(), period{Start: 10, End: 9})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if got := p1ViolationPaths(result); !reflect.DeepEqual(got, []string{"end:periodOrder"}) {
		t.Fatalf("violations = %#v, want object constraint", got)
	}
}

func p1ViolationPaths(result validation.Result) []string {
	violations := result.Violations()
	values := make([]string, 0, len(violations))
	for _, violation := range violations {
		values = append(values, violation.Path()+":"+violation.Code())
	}
	return values
}
