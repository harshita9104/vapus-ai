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

type CacheLog struct {
	managerRequest *pb.DataLogsManagerRequest
	dmStore        *aidmstore.AIStudioDMStore
	logger         *zerolog.Logger
}

func NewCacheLog(managerRequest *pb.DataLogsManagerRequest, dmStore *aidmstore.AIStudioDMStore, logger zerolog.Logger) (*CacheLog, error) {
	return &CacheLog{
		managerRequest: managerRequest,
		dmStore:        dmStore,
		logger:         &logger,
	}, nil
}

func (l *CacheLog) Create(ctx context.Context) error {
	obj := (&models.CacheLogs{}).ConvertFromPb(l.managerRequest.GetSpec().CacheLogs)

	err := l.dmStore.CreateCacheLog(ctx, obj)
	if err != nil {
		l.logger.Error().Err(err).Msg("error while saving api access logs")
		return dmerrors.DMError(apperr.ErrDataLogCreate400, err)
	}

	return nil
}

func (l *CacheLog) GetById() {}
func (l *CacheLog) GetAll()  {}
