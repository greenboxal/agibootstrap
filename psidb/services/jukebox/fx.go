package jukebox

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

var logger = logging.GetLogger("jukebox")

var Module = fx.Module(
	"jukebox",

	inject.ProvideRegisteredService[*ConnectionManager](inject.ServiceRegistrationScopeSingleton, NewConnectionManager),
	inject.ProvideRegisteredService[*QueueManager](inject.ServiceRegistrationScopeSession, NewQueueManager),
)

func init() {
	typesystem.GetType[CommandSheetNode]()
	typesystem.GetType[PlayNoteNode]()
	typesystem.GetType[SetBPMNode]()
	typesystem.GetType[BalanceNode]()
	typesystem.GetType[VolumeNode]()
	typesystem.GetType[PitchBendNode]()

	typesystem.GetType[CommandSheet]()
	typesystem.GetType[NoteCommand]()
	typesystem.GetType[SetBPMCommand]()
	typesystem.GetType[BalanceCommand]()
	typesystem.GetType[VolumeCommand]()
	typesystem.GetType[PitchBendCommand]()
}
