package notification

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrTemplateNotFound,
		ErrInvalidRecipient,
		ErrProviderUnavailable,
		ErrRateLimited,
	}
	for i, a := range errs {
		for j, b := range errs {
			if i != j && errors.Is(a, b) {
				t.Errorf("errors[%d] and errors[%d] should not match", i, j)
			}
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	wrapped := fmt.Errorf("sending: %w", ErrInvalidRecipient)
	if !errors.Is(wrapped, ErrInvalidRecipient) {
		t.Fatal("errors.Is should find ErrInvalidRecipient in wrapped chain")
	}
}
