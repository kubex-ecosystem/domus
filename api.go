// Package kubexdb provides a set of interfaces and types for creating properties, channels, validations, and other components in a Go application.
package kubexdb

import (
	"github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	t "github.com/kubex-ecosystem/domus/internal/types"

	logz "github.com/kubex-ecosystem/logz"
)

type PropertyValBase[T any] interface{ interfaces.IPropertyValBase[T] }
type Property[T any] interface{ interfaces.IProperty[T] }

func NewProperty[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) interfaces.IProperty[T] {
	return t.NewProperty(name, v, withMetrics, cb)
}

type Channel[T any] interface{ interfaces.IChannelCtl[T] }
type ChannelBase[T any] interface{ interfaces.IChannelBase[T] }

func NewChannel[T any](name string, logger *logz.LoggerZ) interfaces.IChannelCtl[T] {
	return t.NewChannelCtl[T](name, logger)
}
func NewChannelCtlWithProperty[T any, P interfaces.IProperty[T]](name string, buffers *int, property P, withMetrics bool, logger *logz.LoggerZ) interfaces.IChannelCtl[T] {
	return t.NewChannelCtlWithProperty[T, P](name, buffers, property, withMetrics, logger)
}
func NewChannelBase[T any](name string, buffers int, logger *logz.LoggerZ) interfaces.IChannelBase[T] {
	return t.NewChannelBase[T](name, buffers, logger)
}

type Validation[T any] interface{ interfaces.IValidation[T] }

func NewValidation[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) interfaces.IValidation[T] {
	return t.NewValidation[T]()
}

type ValidationFunc[T any] interface{ interfaces.IValidationFunc[T] }

func NewValidationFunc[T any](priority int, f func(value *T, args ...any) interfaces.IValidationResult) interfaces.IValidationFunc[T] {
	return t.NewValidationFunc[T](priority, f)
}

type ValidationResult interface{ interfaces.IValidationResult }

func NewValidationResult(isValid bool, message string, metadata map[string]any, err error) interfaces.IValidationResult {
	return t.NewValidationResult(isValid, message, metadata, err)
}

type Environment interface{ interfaces.IEnvironment }

func NewEnvironment(envFile string, isConfidential bool, logger *logz.LoggerZ) (interfaces.IEnvironment, error) {
	return t.NewEnvironment(envFile, isConfidential, logger)
}

type Mapper[T any] interface{ interfaces.IMapper[T] }

func NewMapper[T any](object *T, filePath string) interfaces.IMapper[T] {
	return t.NewMapper[T](object, filePath)
}

type Mutexes interface{ interfaces.IMutexes }

func NewMutexes() interfaces.IMutexes { return t.NewMutexes() }

func NewMutexesType() Mutexes { return t.NewMutexesType() }

type Reference interface{ kbx.Reference }

func NewReference(name string) kbx.Reference { return kbx.NewReference(name) }

type SignalManager[T chan string] interface{ interfaces.ISignalManager[T] }

func NewSignalManager[T chan string](signalChan T, logger *logz.LoggerZ) interfaces.ISignalManager[T] {
	return t.NewSignalManager(signalChan, logger)
}
