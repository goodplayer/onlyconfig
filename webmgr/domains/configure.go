package domains

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/meidoworks/nekoq-component/configure/configapi"

	"github.com/goodplayer/onlyconfig/webmgr/tools"
)

type Datacenter struct {
	DatacenterName        string
	DatacenterDescription string
	TimeCreated           int64
	TimeUpdated           int64
}

type Environment struct {
	EnvName        string
	EnvDescription string
	TimeCreated    int64
	TimeUpdated    int64
}

type Application struct {
	ApplicationId                int64
	ApplicationName              string
	ApplicationDescription       string
	ApplicationOwnerOrganization *Org
	TimeCreated                  int64
	TimeUpdated                  int64
}

type Namespace struct {
	Name        string
	Description string
	Type        string
	OwnerAppId  int64
	TimeCreated int64
	TimeUpdated int64
}

func (n *Namespace) IsOwnerApp(app *Application) bool {
	return n.OwnerAppId == app.ApplicationId
}

type Configure struct {
	ConfigId        int64
	ConfigKey       string
	ConfigNamespace string
	ConfigEnv       string
	ConfigDc        string
	ContentType     string
	Content         string
	ConfigVersion   string
	ConfigStatus    int64
	TimeCreated     int64
	TimeUpdated     int64
}

func (c *Configure) UpdateConfigVersion(seq int64) {
	c.ConfigVersion = tools.VersionToString(seq)
}

func (c *Configure) GenerateSelectorsString(app *Application) string {
	return configapi.SelectorsHelperCacheValue(c.GenerateSelectors(app))
}

func (c *Configure) GenerateSelectors(app *Application) *configapi.Selectors {
	selectors := &configapi.Selectors{
		Data: map[string]string{
			"app": app.ApplicationName,
			"env": c.ConfigEnv,
			"dc":  c.ConfigDc,
		},
	}
	return selectors
}

type ConfigureHandler struct {
	ConfigureRepository  ConfigureRepository
	PushChangeRepository PushChangeRepository
}

func (c *ConfigureHandler) AddEnvDc(ctx context.Context, addType, addName string) error {
	now := time.Now()
	if addType == "dc" {
		if err := c.ConfigureRepository.AddDc(ctx, &Datacenter{
			DatacenterName:        addName,
			DatacenterDescription: addName,
			TimeCreated:           now.UnixMilli(),
			TimeUpdated:           now.UnixMilli(),
		}); errors.Is(err, ErrDuplicatedEnvOrDc) {
			return err
		} else if err != nil {
			return err
		} else {
			return nil
		}
	} else if addType == "env" {
		if err := c.ConfigureRepository.AddEnv(ctx, &Environment{
			EnvName:        addName,
			EnvDescription: addName,
			TimeCreated:    now.UnixMilli(),
			TimeUpdated:    now.UnixMilli(),
		}); errors.Is(err, ErrDuplicatedEnvOrDc) {
			return err
		} else if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return errors.New("unknown addType:" + addType)
	}
}

func (c *ConfigureHandler) LoadEnvAndDcList(ctx context.Context) ([]*Datacenter, []*Environment, error) {
	envs, err := c.ConfigureRepository.LoadEnvList(ctx)
	if err != nil {
		return nil, nil, err
	}
	dcs, err := c.ConfigureRepository.LoadDcList(ctx)
	if err != nil {
		return nil, nil, err
	}
	return dcs, envs, nil
}

type ApplicationItem struct {
	ApplicationId                int64
	ApplicationName              string
	ApplicationDescription       string
	ApplicationOwnerOrganization *Org
	TimeCreated                  int64
	TimeUpdated                  int64

	EnvAndDcList []struct {
		EnvName string
		DcList  []string
	}
}

func (c *ConfigureHandler) LoadApplicationsByOrganizationIdList(ctx context.Context, orgList []*Org) (result []*ApplicationItem, rerr error) {
	for _, org := range orgList {
		applications, err := c.ConfigureRepository.LoadApplicationsByOrganizationId(ctx, org.OrgId)
		if err != nil {
			return nil, err
		}
		for _, app := range applications {
			envAndDcList, err := c.ConfigureRepository.LoadEnvAndDcListByAppId(ctx, app.ApplicationId)
			if err != nil {
				return nil, err
			}
			envAndDcMap := map[string][]string{}
			for _, item := range envAndDcList {
				arr, ok := envAndDcMap[item.EnvName]
				if !ok {
					arr = []string{item.DcName}
				} else {
					arr = append(arr, item.DcName)
				}
				envAndDcMap[item.EnvName] = arr
			}
			var envAndDcResult []struct {
				EnvName string
				DcList  []string
			}
			for key, val := range envAndDcMap {
				envAndDcResult = append(envAndDcResult, struct {
					EnvName string
					DcList  []string
				}{EnvName: key, DcList: val})
			}

			app.ApplicationOwnerOrganization = org
			result = append(result, &ApplicationItem{
				ApplicationId:                app.ApplicationId,
				ApplicationName:              app.ApplicationName,
				ApplicationDescription:       app.ApplicationDescription,
				ApplicationOwnerOrganization: app.ApplicationOwnerOrganization,
				TimeCreated:                  app.TimeCreated,
				TimeUpdated:                  app.TimeUpdated,
				EnvAndDcList:                 envAndDcResult,
			})
		}
	}
	return
}

func (c *ConfigureHandler) CreateApplication(ctx context.Context, org *Org, appName string) error {
	now := time.Now()
	app := &Application{
		ApplicationName:              appName,
		ApplicationDescription:       appName,
		ApplicationOwnerOrganization: org,
		TimeCreated:                  now.UnixMilli(),
		TimeUpdated:                  now.UnixMilli(),
	}

	return c.ConfigureRepository.SaveApplication(ctx, app)
}

func (c *ConfigureHandler) LinkEnvAndDcToApp(ctx context.Context, envName, dcName string, appId int64) error {
	app, err := c.ConfigureRepository.LoadApplicationById(ctx, appId)
	if err != nil {
		return err
	}
	env, err := c.ConfigureRepository.LoadEnvironment(ctx, envName)
	if err != nil {
		return err
	}
	dc, err := c.ConfigureRepository.LoadDatacenter(ctx, dcName)
	if err != nil {
		return err
	}
	has, err := c.ConfigureRepository.ExistsAppEnvDcMapping(ctx, env, dc, app)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	return c.ConfigureRepository.LinkEnvAndDcToApp(ctx, env, dc, app)
}

func (c *ConfigureHandler) AddApplicationNamespace(ctx context.Context, appId int64, nsName string, nsType string) error {
	switch nsType {
	case "public":
	case "application":
	default:
		return errors.New("invalid format of nsType:" + nsType)
	}

	app, err := c.ConfigureRepository.LoadApplicationById(ctx, appId)
	if err != nil {
		return err
	}

	if has, err := c.ConfigureRepository.ExistsApplicationNamespace(ctx, app, nsName); err != nil {
		return err
	} else if has {
		return nil
	}

	now := time.Now()
	ns := &Namespace{
		Name:        nsName,
		Description: nsName,
		Type:        nsType,
		OwnerAppId:  app.ApplicationId,
		TimeCreated: now.UnixMilli(),
		TimeUpdated: now.UnixMilli(),
	}
	return c.ConfigureRepository.AddApplicationNamespace(ctx, app, ns)
}

func (c *ConfigureHandler) QueryApplicationNamespaces(ctx context.Context, appId int64) ([]*Namespace, error) {
	app, err := c.ConfigureRepository.LoadApplicationById(ctx, appId)
	if err != nil {
		return nil, err
	}
	return c.ConfigureRepository.LoadAppNamespaces(ctx, app)
}

type AddConfigurationRequest struct {
	AppId       int64
	Env         string
	Dc          string
	Namespace   string
	Key         string
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

func (c *ConfigureHandler) AddConfiguration(ctx context.Context, req *AddConfigurationRequest) error {
	switch req.ContentType {
	case "general":
	default:
		log.Println("invalid format of content type:", req.ContentType)
		return errors.New("invalid format of content type:" + req.ContentType)
	}

	app, err := c.ConfigureRepository.LoadApplicationById(ctx, req.AppId)
	if err != nil {
		return err
	}
	dc, err := c.ConfigureRepository.LoadDatacenter(ctx, req.Dc)
	if err != nil {
		return err
	}
	env, err := c.ConfigureRepository.LoadEnvironment(ctx, req.Env)
	if err != nil {
		return err
	}
	if exists, err := c.ConfigureRepository.ExistsAppEnvDcMapping(ctx, env, dc, app); err != nil {
		return err
	} else if !exists {
		return errors.New("app-env-dc mapping not exists")
	}
	ns, err := c.ConfigureRepository.LoadNamespace(ctx, req.Namespace)
	if err != nil {
		return err
	}
	if !ns.IsOwnerApp(app) {
		return errors.New("not owner app")
	}
	if exists, err := c.ConfigureRepository.ExistsConfigure(ctx, app, env, dc, ns, req.Key); err != nil {
		return err
	} else if exists {
		return errors.New("configure exists")
	}
	now := time.Now()
	cfg := &Configure{
		ConfigKey:       req.Key,
		ConfigNamespace: ns.Name,
		ConfigEnv:       env.EnvName,
		ConfigDc:        dc.DatacenterName,
		ContentType:     req.ContentType,
		Content:         req.Content,
		ConfigStatus:    0,
		TimeCreated:     now.UnixMilli(),
		TimeUpdated:     now.UnixMilli(),
	}
	if seq, err := c.ConfigureRepository.NextConfigVersionSeq(ctx); err != nil {
		return err
	} else {
		cfg.UpdateConfigVersion(seq)
	}
	if err := c.ConfigureRepository.AddConfiguration(ctx, cfg); err != nil {
		return err
	}
	if err := ApplyConfigureChange(ctx, c.PushChangeRepository, cfg, app); err != nil {
		return err
	}
	return nil
}

func (c *ConfigureHandler) QueryAppConfigList(ctx context.Context, appId int64, env, dc string) (map[string][]*Configure, error) {
	app, err := c.ConfigureRepository.LoadApplicationById(ctx, appId)
	if err != nil {
		return nil, err
	}
	environment, err := c.ConfigureRepository.LoadEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	datacenter, err := c.ConfigureRepository.LoadDatacenter(ctx, dc)
	if err != nil {
		return nil, err
	}
	list, err := c.ConfigureRepository.LoadAppConfigList(ctx, app, environment, datacenter)
	if err != nil {
		return nil, err
	}
	result := map[string][]*Configure{}
	for _, config := range list {
		nsList := result[config.ConfigNamespace]
		nsList = append(nsList, config)
		result[config.ConfigNamespace] = nsList
	}
	return result, nil
}

func (c *ConfigureHandler) QueryConfigureById(ctx context.Context, cfgId int64) (*Configure, error) {
	cfg, err := c.ConfigureRepository.LoadConfigureById(ctx, cfgId)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type UpdateConfigurationRequest struct {
	ConfigId    int64
	ContentType string `json:"ct"`
	Content     string `json:"content"`
}

func (c *ConfigureHandler) UpdateConfigurationById(ctx context.Context, req *UpdateConfigurationRequest) error {
	cfg, err := c.ConfigureRepository.LoadConfigureById(ctx, req.ConfigId)
	if err != nil {
		return err
	}
	switch req.ContentType {
	case "general":
	default:
		log.Println("invalid format of content type:", req.ContentType)
		return errors.New("invalid format of content type:" + req.ContentType)
	}
	cfg.ContentType = req.ContentType
	cfg.Content = req.Content
	if seq, err := c.ConfigureRepository.NextConfigVersionSeq(ctx); err != nil {
		return err
	} else {
		cfg.UpdateConfigVersion(seq)
	}
	now := time.Now()
	cfg.TimeUpdated = now.UnixMilli()
	if err := c.ConfigureRepository.UpdateConfiguration(ctx, cfg); err != nil {
		return err
	}
	ns, err := c.ConfigureRepository.LoadNamespace(ctx, cfg.ConfigNamespace)
	if err != nil {
		return err
	}
	app, err := c.ConfigureRepository.LoadApplicationById(ctx, ns.OwnerAppId)
	if err != nil {
		return err
	}
	if err := ApplyConfigureChange(ctx, c.PushChangeRepository, cfg, app); err != nil {
		return err
	}
	return nil
}
