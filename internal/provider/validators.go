package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// multipleOfValidator validates that an int64 value is a multiple of a given divisor.
type multipleOfValidator struct {
	divisor int64
}

func (v multipleOfValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be a multiple of %d", v.divisor)
}

func (v multipleOfValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v multipleOfValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	val := req.ConfigValue.ValueInt64()
	if val%v.divisor != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			fmt.Sprintf("must be a multiple of %d, got %d", v.divisor, val),
		)
	}
}

// MultipleOf returns a validator that checks an int64 is a multiple of the given divisor.
func MultipleOf(divisor int64) validator.Int64 {
	return multipleOfValidator{divisor: divisor}
}
