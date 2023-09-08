package vm

import (
	"go.uber.org/zap"
)

type Console interface {
	Assert(value any, message *string, args ...any)

	Count(label *string)
	CountReset(label *string)

	Dir(obj any)
	Dirxml(args ...any)

	Group(args ...any)
	GroupCollapsed(args ...any)
	GroupEnd(args ...any)

	Log(args ...any)
	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Exception(args ...any)

	Table(args ...any)
	Time(args ...any)
	TimeEnd(args ...any)
	TimeLog(args ...any)

	Clear()
}

type basicConsole struct {
	logger *zap.SugaredLogger
}

func (b *basicConsole) Assert(value any, message *string, args ...any) {
	panic("implement me")
}

func (b *basicConsole) Count(label *string) {
	panic("implement me")
}

func (b *basicConsole) CountReset(label *string) {
	panic("implement me")
}

func (b *basicConsole) Dir(obj any) {
	b.Log(obj)
}

func (b *basicConsole) Dirxml(args ...any) {
	panic("implement me")
}

func (b *basicConsole) Group(args ...any) {
	panic("implement me")
}

func (b *basicConsole) GroupCollapsed(args ...any) {
	panic("implement me")
}

func (b *basicConsole) GroupEnd(args ...any) {
	panic("implement me")
}

func (b *basicConsole) Log(args ...any) {
	b.logger.Info(args...)
}

func (b *basicConsole) Trace(args ...any) {
	b.logger.Debug(args...)
}

func (b *basicConsole) Debug(args ...any) {
	b.logger.Debug(args...)
}

func (b *basicConsole) Info(args ...any) {
	b.logger.Info(args...)
}

func (b *basicConsole) Warn(args ...any) {
	b.logger.Warn(args...)
}

func (b *basicConsole) Error(args ...any) {
	b.logger.Error(args...)
}

func (b *basicConsole) Exception(args ...any) {
	b.Error(args...)
}

func (b *basicConsole) Table(args ...any) {
	b.logger.Info(args...)
}

func (b *basicConsole) Time(args ...any) {
	panic("implement me")
}

func (b *basicConsole) TimeEnd(args ...any) {
	panic("implement me")
}

func (b *basicConsole) TimeLog(args ...any) {
	panic("implement me")
}

func (b *basicConsole) Clear() {
}
