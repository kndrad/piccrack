package cmd

import (
	"fmt"
)

func OnShutdown(funcs ...func() error) func() error {
	return func() error {
		for _, f := range funcs {
			if err := f(); err != nil {
				return fmt.Errorf("executing func: %w", err)
			}
		}

		return nil
	}
}
