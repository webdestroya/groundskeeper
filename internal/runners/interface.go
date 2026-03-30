package runners

import "context"

type Runner interface {
	Run(context.Context) error
}
