package services

import (
	"context"

	"github.com/rs/zerolog"
	pb "github.com/vapusdata-ecosystem/apis/protos/vapusai-studio/v1alpha1"
	aidmstore "github.com/vapusdata-ecosystem/vapusai/core/app/datarepo/aistudio"
	apperr "github.com/vapusdata-ecosystem/vapusai/core/app/errors"
	"github.com/vapusdata-ecosystem/vapusai/core/models"
	dmerrors "github.com/vapusdata-ecosystem/vapusai/core/pkgs/errors"
)

type APIAccessLog struct {
	managerRequest *pb.DataLogsManagerRequest
	dmStore        *aidmstore.AIStudioDMStore
	logger         *zerolog.Logger
}

func NewAPIAccessLog(managerRequest *pb.DataLogsManagerRequest, dmStore *aidmstore.AIStudioDMStore, logger zerolog.Logger) (*APIAccessLog, error) {
	return &APIAccessLog{
		managerRequest: managerRequest,
		dmStore:        dmStore,
		logger:         &logger,
	}, nil
}

func (l *APIAccessLog) Create(ctx context.Context) error {
	obj := (&models.APIAccessLogs{}).ConvertFromPb(l.managerRequest.GetSpec().ApiAccessLogs)

	err := l.dmStore.CreateAPIAccessLog(ctx, obj)
	if err != nil {
		l.logger.Error().Err(err).Msg("error while saving api access logs")
		return dmerrors.DMError(apperr.ErrDataLogCreate400, err)
	}

	return nil
}

func (l *APIAccessLog) GetById() {}
func (l *APIAccessLog) GetAll()  {}
