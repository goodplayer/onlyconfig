package domains

import (
	"context"
	"errors"
)

var ErrDuplicatedEnvOrDc = errors.New("duplicated env or dc in repository")

type ConfigureRepository interface {
	AddDc(ctx context.Context, dc *Datacenter) error
	AddEnv(ctx context.Context, env *Environment) error
	SaveApplication(ctx context.Context, application *Application) error
	LinkEnvAndDcToApp(ctx context.Context, env *Environment, dc *Datacenter, app *Application) error
	AddApplicationNamespace(ctx context.Context, app *Application, ns *Namespace) error
	AddConfiguration(ctx context.Context, cfg *Configure) error
	UpdateConfiguration(ctx context.Context, cfg *Configure) error
	NextConfigVersionSeq(ctx context.Context) (int64, error)

	LoadDcList(ctx context.Context) ([]*Datacenter, error)
	LoadEnvList(ctx context.Context) ([]*Environment, error)
	LoadApplicationsByOrganizationId(ctx context.Context, orgId string) ([]*Application, error)
	LoadApplicationById(ctx context.Context, applicationId int64) (*Application, error)
	LoadDatacenter(ctx context.Context, dc string) (*Datacenter, error)
	LoadEnvironment(ctx context.Context, env string) (*Environment, error)
	ExistsAppEnvDcMapping(ctx context.Context, env *Environment, dc *Datacenter, app *Application) (bool, error)
	LoadEnvAndDcListByAppId(ctx context.Context, appId int64) ([]struct {
		EnvName string
		DcName  string
	}, error)
	ExistsApplicationNamespace(ctx context.Context, app *Application, nsName string) (bool, error)
	LoadAppNamespaces(ctx context.Context, app *Application) ([]*Namespace, error)
	LoadNamespace(ctx context.Context, nsName string) (*Namespace, error)
	ExistsConfigure(ctx context.Context, app *Application, env *Environment, dc *Datacenter, ns *Namespace, key string) (bool, error)
	LoadAppConfigList(ctx context.Context, app *Application, env *Environment, dc *Datacenter) ([]*Configure, error)
	LoadConfigureById(ctx context.Context, cfgId int64) (*Configure, error)
}
