package dmcontrollers

import (
	"context"

	"github.com/rs/zerolog"
	mpb "github.com/vapusdata-ecosystem/apis/protos/models/v1alpha1"
	pb "github.com/vapusdata-ecosystem/apis/protos/vapusai-studio/v1alpha1"
	pkgs "github.com/vapusdata-ecosystem/vapusai/aistudio/pkgs"
	"github.com/vapusdata-ecosystem/vapusai/aistudio/services"
	dmsvc "github.com/vapusdata-ecosystem/vapusai/aistudio/services"
	pbtools "github.com/vapusdata-ecosystem/vapusai/core/pkgs/pbtools"
	dmutils "github.com/vapusdata-ecosystem/vapusai/core/pkgs/utils"
)

type VapusDataLogs struct {
	pb.UnimplementedDataLogsServer
	validator  *dmutils.DMValidator
	DMServices *dmsvc.AIStudioServices
	logger     zerolog.Logger
}

var VapusDataLogsManager *VapusDataLogs

func NewVapusDataLogs() *VapusDataLogs {
	l := pkgs.GetSubDMLogger(pkgs.CNTRLR, "VapusDataLogs")
	validator, err := dmutils.NewDMValidator()
	if err != nil {
		l.Panic().Err(err).Msg("Error while loading validator")
	}

	l.Info().Msg("VapusDataLogs Controller initialized")
	return &VapusDataLogs{
		validator:  validator,
		logger:     l,
		DMServices: dmsvc.AIStudioServiceManager,
	}
}

func InitVapusDataLogs() {
	if VapusDataLogsManager == nil {
		VapusDataLogsManager = NewVapusDataLogs()
	}
}

func (v *VapusDataLogs) Create(ctx context.Context, request *pb.DataLogsManagerRequest) (*mpb.VapusCreateResponse, error) {
	v.logger.Info().Msg("Saving data log...")

	agent, err := v.DMServices.NewDataLogsAgent(ctx, services.WithDataLogsManagerRequest(request), services.WithDataLogsManagerAction(mpb.ResourceLcActions_CREATE))
	if err != nil {
		v.logger.Error().Err(err).Msg("Error while creating data log manager request")
		return nil, err
	}
	defer func() {
		dmutils.CleanPointers(agent)
	}()
	err = agent.Act(ctx)
	if err != nil {
		v.logger.Error().Err(err).Msg("Error while processing data log creation request")
		return nil, err
	}
	response := agent.GetCreateResponse()
	response.DmResp = pbtools.HandleDMResponse(ctx, "Data log create action executed successfully", "200")
	return response, nil
}

func (dmc *VapusDataLogs) GetById(ctx context.Context, request *pb.GetByIdDataLogRequest) (*pb.GetByIdDataLogResponse, error) {
	dmc.logger.Info().Msg("Saving data log...")
	return nil, nil
	//return dmc.DMServices.SaveAPIAccessLog(ctx, request)
}

func (dmc *VapusDataLogs) GetAll(ctx context.Context, request *pb.GetAllDataLogRequest) (*pb.GetAllDataLogResponse, error) {
	dmc.logger.Info().Msg("Saving data log...")
	return nil, nil
	//return dmc.DMServices.SaveAPIAccessLog(ctx, request)
}
