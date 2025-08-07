package services

import (
	"context"

	mpb "github.com/vapusdata-ecosystem/apis/protos/models/v1alpha1"
	pb "github.com/vapusdata-ecosystem/apis/protos/vapusai-studio/v1alpha1"
	pkgs "github.com/vapusdata-ecosystem/vapusai/aistudio/pkgs"
	aidmstore "github.com/vapusdata-ecosystem/vapusai/core/app/datarepo/aistudio"
	apperr "github.com/vapusdata-ecosystem/vapusai/core/app/errors"
	encryption "github.com/vapusdata-ecosystem/vapusai/core/pkgs/encryption"
	dmerrors "github.com/vapusdata-ecosystem/vapusai/core/pkgs/errors"
	dmutils "github.com/vapusdata-ecosystem/vapusai/core/pkgs/utils"
	processes "github.com/vapusdata-ecosystem/vapusai/core/process"
	types "github.com/vapusdata-ecosystem/vapusai/core/types"
)

type DataLogsAgent struct {
	*processes.VapusInterfaceBase
	managerRequest *pb.DataLogsManagerRequest
	dmStore        *aidmstore.AIStudioDMStore
}

type DataLogsIntAgentOpts func(*DataLogsAgent)

func WithDataLogsManagerRequest(managerRequest *pb.DataLogsManagerRequest) DataLogsIntAgentOpts {
	return func(v *DataLogsAgent) {
		v.managerRequest = managerRequest
	}
}

func WithDataLogsManagerAction(action mpb.ResourceLcActions) DataLogsIntAgentOpts {
	return func(v *DataLogsAgent) {
		v.Action = action.String()
	}
}

func (s *AIStudioServices) NewDataLogsAgent(ctx context.Context, opts ...DataLogsIntAgentOpts) (*DataLogsAgent, error) {
	vapusPlatformClaim, ok := encryption.GetCtxClaim(ctx)
	if !ok {
		s.Logger.Error().Ctx(ctx).Msg("error while getting claim metadata from context")
		return nil, dmerrors.DMError(encryption.ErrInvalidJWTClaims, nil)
	}
	datalog := &DataLogsAgent{
		dmStore: s.DMStore,
		VapusInterfaceBase: &processes.VapusInterfaceBase{
			CtxClaim: vapusPlatformClaim,
			InitAt:   dmutils.GetEpochTime(),
		},
	}
	for _, opt := range opts {
		opt(datalog)
	}
	datalog.SetAgentId()
	datalog.Logger = pkgs.GetSubDMLogger(types.DATALOGAGENT.String(), datalog.AgentId)
	return datalog, nil
}

func (d *DataLogsAgent) Act(ctx context.Context) error {
	switch d.GetAction() {
	case mpb.ResourceLcActions_CREATE.String():
		return d.Create(ctx)
	default:
		d.Logger.Error().Msg("invalid action for DataLogsAgent")
		return dmerrors.DMError(apperr.ErrInvalidAction, nil)
	}
}

func (d *DataLogsAgent) Create(ctx context.Context) error {
	logType := d.managerRequest.LogType

	service, err := d.getLogService(logType)
	if err != nil {
		d.Logger.Error().Msg("invalid data log type")
		return err
	}

	err = service.Create(ctx)
	if err != nil {
		d.Logger.Error().Msg("invalid data log type")
		return err
	}
	return nil
}

func (d *DataLogsAgent) getLogService(logType mpb.DataLogType) (IDataLog, error) {
	switch logType {
	case 0:
		return NewAPIAccessLog(d.managerRequest, d.dmStore, d.Logger)
	case 1:
		return NewCacheLog(d.managerRequest, d.dmStore, d.Logger)
	}

	return nil, dmerrors.DMError(apperr.ErrDataLogInvalidType, nil)
}
