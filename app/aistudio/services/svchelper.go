package services

import (
	"context"
	"encoding/json"

	mpb "github.com/vapusdata-ecosystem/apis/protos/models/v1alpha1"
	pkgs "github.com/vapusdata-ecosystem/vapusai/aistudio/pkgs"
	aidmstore "github.com/vapusdata-ecosystem/vapusai/core/app/datarepo/aistudio"
	apperr "github.com/vapusdata-ecosystem/vapusai/core/app/errors"
	apppkgs "github.com/vapusdata-ecosystem/vapusai/core/app/pkgs"
	"github.com/vapusdata-ecosystem/vapusai/core/models"
	encryption "github.com/vapusdata-ecosystem/vapusai/core/pkgs/encryption"
	dmutils "github.com/vapusdata-ecosystem/vapusai/core/pkgs/utils"
)

func getOrganizationArtifactCreds(ctx context.Context, organization *models.Organization, store *aidmstore.AIStudioDMStore) *models.OCILoginCreds {
	if organization.ArtifactStorage == nil || organization.ArtifactStorage.NetParams == nil || len(organization.ArtifactStorage.NetParams.DsCreds) < 1 {
		helperLogger.Err(apperr.ErrOrganizationArtifactStore404).Msg("error while fetching organization artifact store")
		return nil
	}
	helperLogger.Debug().Msgf("reading artifact store cred with secret name - %v", organization.ArtifactStorage.NetParams.DsCreds[0].SecretName)
	creds, err := store.GetDataCredsFromSecret(ctx, organization.ArtifactStorage.NetParams.DsCreds[0].SecretName)
	if err != nil {
		helperLogger.Err(err).Msgf("error while fetching organization artifact store")
		return nil
	}
	helperLogger.Info().Msg("organization artifact store fetched successfully")
	helperLogger.Info().Msgf("organization artifact store creds fetched successfully  - %v", organization.ArtifactStorage.NetParams.Address)
	helperLogger.Info().Msgf("organization artifact store creds fetched successfully++++++++  - %v - %v", creds.Credentials.Username, creds.Credentials.Password)
	return &models.OCILoginCreds{
		Username: creds.Credentials.Username,
		Password: creds.Credentials.Password,
		URL:      organization.ArtifactStorage.NetParams.Address,
	}
}

func setJWTAuthzParams(ctx context.Context, secretName string, secretStoreClient *apppkgs.SecretStore, usePlatform bool, jwtparam *models.JWTParams) (*models.JWTParams, error) {
	if jwtparam == nil || jwtparam.PublicJwtKey == "" || jwtparam.PrivateJwtKey == "" {
		usePlatform = true
		helperLogger.Info().Msg("using platform jwt secrets because jwt secrets are not provided in request")
	}
	var err error
	if usePlatform {
		err = secretStoreClient.WriteSecret(ctx, pkgs.SvcPackageManager.VapusJwtAuth.Opts, secretName)
		if err != nil {
			helperLogger.Err(err).Msgf("error while swapping default platform JWT keys for given resource - %v", secretName)
			return nil, err
		}
	} else {
		err = secretStoreClient.WriteSecret(ctx, &encryption.JWTAuthn{
			PublicJWTKey:     jwtparam.PublicJwtKey,
			PrivateJWTKey:    jwtparam.PrivateJwtKey,
			SigningAlgorithm: jwtparam.SigningAlgorithm,
		}, secretName)
		if err != nil {
			helperLogger.Err(err).Msgf("error while swapping JWT keys for given resource - %v", secretName)
			return nil, err
		}
	}
	return &models.JWTParams{
		VId:                 secretName,
		Name:                secretName,
		SigningAlgorithm:    pkgs.SvcPackageManager.VapusJwtAuth.Opts.SigningAlgorithm,
		IsAlreadyInSecretBs: true,
		Status:              mpb.CommonStatus_ACTIVE.String(),
	}, nil
}

func getsecretPassCode(resource, resourceId string) string {
	return dmutils.SlugifyBase(resource) + "_" + dmutils.SlugifyBase(resourceId)
}

func getOrganizationAuthn(ctx context.Context, organization *models.Organization, store *aidmstore.AIStudioDMStore, forSignValidation bool) (*encryption.JWTAuthn, error) {
	authnObj := organization.GetAuthnJwtParams()
	helperLogger.Info().Msgf("authnObj - %v", authnObj)
	secretStr, err := store.SecretStore.ReadSecret(ctx, authnObj.GetName())
	if err != nil {
		helperLogger.Err(err).Msgf("error while fetching organization authn secrets")
		return nil, err
	}
	helperLogger.Info().Msgf("authnObj - %v", secretStr)
	jwtParams := &encryption.JWTAuthn{}
	err = json.Unmarshal([]byte(dmutils.AnyToStr(secretStr)), jwtParams)
	if err != nil {
		helperLogger.Err(err).Ctx(ctx).Msg("error while unmarshaling the organization jwt from secret store")
		return nil, err
	}

	if forSignValidation {
		jwtParams.PrivateJWTKey = ""
		jwtParams.ForPublicValidation = true
	}
	return jwtParams, nil
}
