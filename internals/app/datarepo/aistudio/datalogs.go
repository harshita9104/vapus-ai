package aidmstore

import (
	"context"

	apppkgs "github.com/vapusdata-ecosystem/vapusai/core/app/pkgs"
	"github.com/vapusdata-ecosystem/vapusai/core/models"
)

func (ds *AIStudioDMStore) CreateAPIAccessLog(ctx context.Context, obj *models.APIAccessLogs) error {
	_, err := ds.Db.PostgresClient.DB.NewInsert().Model(obj).ModelTableExpr(apppkgs.APIAccessLogTable).Exec(ctx)
	if err != nil {
		ds.logger.Err(err).Ctx(ctx).Msg("error while saving api access log to datastore")
		return err
	}

	return nil
}

func (ds *AIStudioDMStore) CreateCacheLog(ctx context.Context, obj *models.APIAccessLogs) error {
	_, err := ds.Db.PostgresClient.DB.NewInsert().Model(obj).ModelTableExpr(apppkgs.CacheLogTable).Exec(ctx)
	if err != nil {
		ds.logger.Err(err).Ctx(ctx).Msg("error while saving cache log to datastore")
		return err
	}

	return nil
}
