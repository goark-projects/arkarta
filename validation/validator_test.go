package validation

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDefaultValidatorValidatesStructTags(t *testing.T) {
	t.Parallel()

	type address struct {
		City string `json:"city" arkarta:"required,notblank"`
	}
	type user struct {
		Name    string  `json:"name" arkarta:"required,min=2,max=8"`
		Email   string  `json:"email" arkarta:"email"`
		Status  string  `json:"status" arkarta:"oneof=ACTIVE|DISABLED"`
		Age     int     `json:"age" arkarta:"min=18,max=120"`
		Address address `json:"address"`
	}

	result, err := Validate(context.Background(), user{
		Name:   "张",
		Email:  "bad",
		Status: "LOCKED",
		Age:    12,
		Address: address{
			City: " ",
		},
	})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Valid() {
		t.Fatal("result should be invalid")
	}
	got := violationPaths(result)
	want := []string{"name:min", "email:email", "status:oneof", "age:min", "address.city:notblank"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %#v, want %#v", got, want)
	}
	if result.Error() == nil {
		t.Fatal("invalid result should expose error")
	}
}

func TestDefaultValidatorValidatesNestedSlices(t *testing.T) {
	t.Parallel()

	type item struct {
		SKU string `json:"sku" arkarta:"required"`
	}
	type order struct {
		Items []item `json:"items"`
		Tags  []string
	}

	result, err := Validate(context.Background(), order{Items: []item{{}, {SKU: "ok"}}, Tags: []string{"a"}})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	got := violationPaths(result)
	want := []string{"items[0].sku:required"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %#v, want %#v", got, want)
	}
}

func TestDefaultValidatorAllowsNilForOptionalRules(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name   *string `json:"name" arkarta:"min=2"`
		Status *string `json:"status" arkarta:"oneof=ACTIVE|DISABLED"`
	}
	result, err := Validate(context.Background(), payload{})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("violations = %#v, want none", result.Violations())
	}
}

func TestDefaultValidatorCustomConstraintAndContext(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name" arkarta:"ark"`
	}
	validator := NewValidator(WithConstraint(ConstraintFunc{
		RuleName: "ark",
		Fn: func(ctx context.Context, field FieldContext) (Violation, bool, error) {
			if err := ctx.Err(); err != nil {
				return Violation{}, false, err
			}
			if field.Value().String() != "arkarta" {
				return NewViolation(field.Path(), "ark", "必须等于 arkarta", field.Value().String()), true, nil
			}
			return Violation{}, false, nil
		},
	}))
	result, err := validator.Validate(context.Background(), payload{Name: "goark"})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if got := violationPaths(result); !reflect.DeepEqual(got, []string{"name:ark"}) {
		t.Fatalf("violations = %#v", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := validator.Validate(ctx, payload{Name: "arkarta"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err = %v, want context.Canceled", err)
	}
}

func TestDefaultValidatorRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	if _, err := (*DefaultValidator)(nil).Validate(context.Background(), struct{}{}); !errors.Is(err, ErrNilValidator) {
		t.Fatalf("nil validator err = %v, want ErrNilValidator", err)
	}
	if _, err := Validate(context.Background(), nil); !errors.Is(err, ErrNilValue) {
		t.Fatalf("nil value err = %v, want ErrNilValue", err)
	}
	type payload struct {
		Name string `arkarta:"min=x"`
	}
	if _, err := Validate(context.Background(), payload{Name: "a"}); !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("invalid rule err = %v, want ErrInvalidRule", err)
	}
}

func violationPaths(result Result) []string {
	violations := result.Violations()
	values := make([]string, 0, len(violations))
	for _, violation := range violations {
		values = append(values, violation.Path()+":"+violation.Code())
	}
	return values
}
