package postgres

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/dbtxn"
)

type Environment struct {
	EnvName     string `xorm:"'env_name' pk"`
	EnvDesc     string `xorm:"'env_description'"`
	TimeCreated int64  `xorm:"'time_created'"`
	TimeUpdated int64  `xorm:"'time_updated'"`
}

func (u *Environment) TableName() string {
	return "onlyconfig_environment"
}

type Datacenter struct {
	DcName      string `xorm:"'datacenter_name' pk"`
	DcDesc      string `xorm:"'datacenter_description'"`
	TimeCreated int64  `xorm:"'time_created'"`
	TimeUpdated int64  `xorm:"'time_updated'"`
}

func (u *Datacenter) TableName() string {
	return "onlyconfig_datacenter"
}

type Application struct {
	AppId         int64  `xorm:"'application_id' pk autoincr"`
	AppName       string `xorm:"'application_name'"`
	AppDesc       string `xorm:"'application_description'"`
	AppOwnerOrgId string `xorm:"'application_owner_org'"`
	TimeCreated   int64  `xorm:"'time_created'"`
	TimeUpdated   int64  `xorm:"'time_updated'"`
}

func (u *Application) TableName() string {
	return "onlyconfig_application"
}

type AppDetail struct {
	DetailId    int64  `xorm:"'application_detail_id' pk autoincr"`
	AppId       int64  `xorm:"'application_id'"`
	EnvName     string `xorm:"'env_name'"`
	DcName      string `xorm:"'datacenter_name'"`
	TimeCreated int64  `xorm:"'time_created'"`
	TimeUpdated int64  `xorm:"'time_updated'"`
}

func (u *AppDetail) TableName() string {
	return "onlyconfig_application_detail"
}

type Namespace struct {
	Name        string `xorm:"'namespace_name' pk"`
	Description string `xorm:"'namespace_description'"`
	Type        string `xorm:"'namespace_type'"`
	AppId       int64  `xorm:"'namespace_app'"`
	TimeCreated int64  `xorm:"'time_created'"`
	TimeUpdated int64  `xorm:"'time_updated'"`
}

func (u *Namespace) TableName() string {
	return "onlyconfig_namespace"
}

type Configure struct {
	ConfigId        int64  `xorm:"'config_id' pk autoincr"`
	ConfigKey       string `xorm:"'config_key'"`
	ConfigNamespace string `xorm:"'config_namespace'"`
	ConfigEnv       string `xorm:"'config_env'"`
	ConfigDc        string `xorm:"'config_datacenter'"`
	ContentType     string `xorm:"'config_content_type'"`
	Content         string `xorm:"'config_content'"`
	ConfigVersion   string `xorm:"'config_version' default ''"`
	ConfigStatus    int64  `xorm:"'config_status'"`
	TimeCreated     int64  `xorm:"'time_created'"`
	TimeUpdated     int64  `xorm:"'time_updated'"`
}

func (u *Configure) TableName() string {
	return "onlyconfig_config"
}

type ConfigureStoreImpl struct {
}

func (c *ConfigureStoreImpl) AddDc(ctx context.Context, dc *domains.Datacenter) error {
	sess := dbtxn.GetTxn(ctx)

	if has, err := sess.Where("datacenter_name = ?", dc.DatacenterName).Get(new(Datacenter)); err != nil {
		return err
	} else if has {
		return domains.ErrDuplicatedEnvOrDc
	}

	_, err := sess.Insert(&Datacenter{
		DcName:      dc.DatacenterName,
		DcDesc:      dc.DatacenterDescription,
		TimeCreated: dc.TimeCreated,
		TimeUpdated: dc.TimeUpdated,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigureStoreImpl) AddEnv(ctx context.Context, env *domains.Environment) error {
	sess := dbtxn.GetTxn(ctx)

	if has, err := sess.Where("env_name = ?", env.EnvName).Get(new(Environment)); err != nil {
		return err
	} else if has {
		return domains.ErrDuplicatedEnvOrDc
	}

	_, err := sess.Insert(&Environment{
		EnvName:     env.EnvName,
		EnvDesc:     env.EnvDescription,
		TimeCreated: env.TimeCreated,
		TimeUpdated: env.TimeUpdated,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigureStoreImpl) LoadDcList(ctx context.Context) (result []*domains.Datacenter, rerr error) {
	var dcList []*Datacenter

	sess := dbtxn.GetTxn(ctx)
	if err := sess.OrderBy("time_created").Find(&dcList); err != nil {
		return nil, err
	}

	for _, dc := range dcList {
		result = append(result, &domains.Datacenter{
			DatacenterName:        dc.DcName,
			DatacenterDescription: dc.DcDesc,
			TimeCreated:           dc.TimeCreated,
			TimeUpdated:           dc.TimeUpdated,
		})
	}

	return
}

func (c *ConfigureStoreImpl) LoadEnvList(ctx context.Context) (result []*domains.Environment, rerr error) {
	var envList []*Environment

	sess := dbtxn.GetTxn(ctx)
	if err := sess.OrderBy("time_created").Find(&envList); err != nil {
		return nil, err
	}

	for _, env := range envList {
		result = append(result, &domains.Environment{
			EnvName:        env.EnvName,
			EnvDescription: env.EnvDesc,
			TimeCreated:    env.TimeCreated,
			TimeUpdated:    env.TimeUpdated,
		})
	}

	return
}

func (c *ConfigureStoreImpl) LoadApplicationsByOrganizationId(ctx context.Context, orgId string) (result []*domains.Application, rerr error) {
	var appList []*Application

	sess := dbtxn.GetTxn(ctx)
	if err := sess.Where("application_owner_org = ?", orgId).Find(&appList); err != nil {
		return nil, err
	}

	for _, app := range appList {
		result = append(result, &domains.Application{
			ApplicationId:                app.AppId,
			ApplicationName:              app.AppName,
			ApplicationDescription:       app.AppDesc,
			ApplicationOwnerOrganization: nil, //FIXME fill in the owner org
			TimeCreated:                  app.TimeCreated,
			TimeUpdated:                  app.TimeUpdated,
		})
	}

	return
}

func (c *ConfigureStoreImpl) SaveApplication(ctx context.Context, application *domains.Application) error {
	app := &Application{
		AppName:       application.ApplicationName,
		AppDesc:       application.ApplicationDescription,
		AppOwnerOrgId: application.ApplicationOwnerOrganization.OrgId,
		TimeCreated:   application.TimeCreated,
		TimeUpdated:   application.TimeUpdated,
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(app); err != nil {
		return err
	}
	return nil
}

func (c *ConfigureStoreImpl) LinkEnvAndDcToApp(ctx context.Context, env *domains.Environment, dc *domains.Datacenter, app *domains.Application) error {
	now := time.Now()
	detail := &AppDetail{
		AppId:       app.ApplicationId,
		EnvName:     env.EnvName,
		DcName:      dc.DatacenterName,
		TimeCreated: now.UnixMilli(),
		TimeUpdated: now.UnixMilli(),
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(detail); err != nil {
		return err
	}
	return nil
}

func (c *ConfigureStoreImpl) LoadApplicationById(ctx context.Context, applicationId int64) (*domains.Application, error) {
	app := new(Application)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("application_id = ?", applicationId).Get(app); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("application not found")
	} else {
		return &domains.Application{
			ApplicationId:                app.AppId,
			ApplicationName:              app.AppName,
			ApplicationDescription:       app.AppDesc,
			ApplicationOwnerOrganization: nil, //FIXME fill in the owner org
			TimeCreated:                  app.TimeCreated,
			TimeUpdated:                  app.TimeUpdated,
		}, nil
	}
}

func (c *ConfigureStoreImpl) LoadDatacenter(ctx context.Context, dc string) (*domains.Datacenter, error) {
	datacenter := new(Datacenter)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("datacenter_name = ?", dc).Get(datacenter); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("datacenter not found")
	} else {
		return &domains.Datacenter{
			DatacenterName:        datacenter.DcName,
			DatacenterDescription: datacenter.DcDesc,
			TimeCreated:           datacenter.TimeCreated,
			TimeUpdated:           datacenter.TimeUpdated,
		}, nil
	}
}

func (c *ConfigureStoreImpl) LoadEnvironment(ctx context.Context, env string) (*domains.Environment, error) {
	environment := new(Environment)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("env_name = ?", env).Get(environment); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("environment not found")
	} else {
		return &domains.Environment{
			EnvName:        environment.EnvName,
			EnvDescription: environment.EnvDesc,
			TimeCreated:    environment.TimeCreated,
			TimeUpdated:    environment.TimeUpdated,
		}, nil
	}
}

func (c *ConfigureStoreImpl) ExistsAppEnvDcMapping(ctx context.Context, env *domains.Environment, dc *domains.Datacenter, app *domains.Application) (bool, error) {
	detail := new(AppDetail)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("env_name = ? and datacenter_name = ? and application_id = ?", env.EnvName, dc.DatacenterName, app.ApplicationId).Get(detail); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (c *ConfigureStoreImpl) LoadEnvAndDcListByAppId(ctx context.Context, appId int64) (r []struct {
	EnvName string
	DcName  string
}, rerr error) {
	var result []*AppDetail
	sess := dbtxn.GetTxn(ctx)
	if err := sess.Where("application_id = ?", appId).Find(&result); err != nil {
		return nil, err
	}
	for _, app := range result {
		r = append(r, struct {
			EnvName string
			DcName  string
		}{EnvName: app.EnvName, DcName: app.DcName})
	}
	return
}

func (c *ConfigureStoreImpl) ExistsApplicationNamespace(ctx context.Context, app *domains.Application, nsName string) (bool, error) {
	ns := new(Namespace)

	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("namespace_app = ? and namespace_name = ?", app.ApplicationId, nsName).Get(ns); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (c *ConfigureStoreImpl) AddApplicationNamespace(ctx context.Context, app *domains.Application, ns *domains.Namespace) error {
	namespace := &Namespace{
		Name:        ns.Name,
		Description: ns.Description,
		Type:        ns.Type,
		AppId:       app.ApplicationId,
		TimeCreated: ns.TimeCreated,
		TimeUpdated: ns.TimeUpdated,
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(namespace); err != nil {
		return err
	}
	return nil
}

func (c *ConfigureStoreImpl) LoadAppNamespaces(ctx context.Context, app *domains.Application) (result []*domains.Namespace, rerr error) {
	var r []*Namespace
	sess := dbtxn.GetTxn(ctx)
	if err := sess.Where("namespace_app = ?", app.ApplicationId).Find(&r); err != nil {
		return nil, err
	}
	for _, ns := range r {
		result = append(result, &domains.Namespace{
			Name:        ns.Name,
			Description: ns.Description,
			Type:        ns.Type,
			OwnerAppId:  ns.AppId,
			TimeCreated: ns.TimeCreated,
			TimeUpdated: ns.TimeUpdated,
		})
	}
	return
}

func (c *ConfigureStoreImpl) AddConfiguration(ctx context.Context, cfg *domains.Configure) error {
	configure := &Configure{
		ConfigKey:       cfg.ConfigKey,
		ConfigNamespace: cfg.ConfigNamespace,
		ConfigEnv:       cfg.ConfigEnv,
		ConfigDc:        cfg.ConfigDc,
		ContentType:     cfg.ContentType,
		Content:         cfg.Content,
		ConfigStatus:    cfg.ConfigStatus,
		ConfigVersion:   cfg.ConfigVersion,
		TimeCreated:     cfg.TimeCreated,
		TimeUpdated:     cfg.TimeUpdated,
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.Insert(configure); err != nil {
		return err
	}
	return nil
}

func (c *ConfigureStoreImpl) LoadNamespace(ctx context.Context, nsName string) (*domains.Namespace, error) {
	ns := new(Namespace)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("namespace_name = ?", nsName).Get(ns); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("namespace not found")
	}
	return &domains.Namespace{
		Name:        ns.Name,
		Description: ns.Description,
		Type:        ns.Type,
		OwnerAppId:  ns.AppId,
		TimeCreated: ns.TimeCreated,
		TimeUpdated: ns.TimeUpdated,
	}, nil
}

func (c *ConfigureStoreImpl) ExistsConfigure(ctx context.Context, app *domains.Application, env *domains.Environment, dc *domains.Datacenter, ns *domains.Namespace, key string) (bool, error) {
	cfg := new(Configure)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("config_key = ? and config_namespace = ? and config_env = ? and config_datacenter = ?", key, ns.Name, env.EnvName, dc.DatacenterName).Get(cfg); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (c *ConfigureStoreImpl) LoadAppConfigList(ctx context.Context, app *domains.Application, env *domains.Environment, dc *domains.Datacenter) (result []*domains.Configure, rerr error) {
	sess := dbtxn.GetTxn(ctx)
	var namespaces []*Namespace
	if err := sess.Where("namespace_app = ?", app.ApplicationId).Find(&namespaces); err != nil {
		return nil, err
	}
	var namespaceNames []string
	for _, ns := range namespaces {
		namespaceNames = append(namespaceNames, ns.Name)
	}
	var list []*Configure
	if err := sess.Where("config_env = ? and config_datacenter = ?", env.EnvName, dc.DatacenterName).In("config_namespace", namespaceNames).Find(&list); err != nil {
		return nil, err
	}
	for _, cfg := range list {
		result = append(result, &domains.Configure{
			ConfigId:        cfg.ConfigId,
			ConfigKey:       cfg.ConfigKey,
			ConfigNamespace: cfg.ConfigNamespace,
			ConfigEnv:       cfg.ConfigEnv,
			ConfigDc:        cfg.ConfigDc,
			ContentType:     cfg.ContentType,
			Content:         cfg.Content,
			ConfigVersion:   cfg.ConfigVersion,
			ConfigStatus:    cfg.ConfigStatus,
			TimeCreated:     cfg.TimeCreated,
			TimeUpdated:     cfg.TimeUpdated,
		})
	}
	return
}

func (c *ConfigureStoreImpl) LoadConfigureById(ctx context.Context, cfgId int64) (*domains.Configure, error) {
	cfg := new(Configure)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("config_id = ?", cfgId).Get(cfg); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("config not found")
	}
	return &domains.Configure{
		ConfigId:        cfg.ConfigId,
		ConfigKey:       cfg.ConfigKey,
		ConfigNamespace: cfg.ConfigNamespace,
		ConfigEnv:       cfg.ConfigEnv,
		ConfigDc:        cfg.ConfigDc,
		ContentType:     cfg.ContentType,
		Content:         cfg.Content,
		ConfigVersion:   cfg.ConfigVersion,
		ConfigStatus:    cfg.ConfigStatus,
		TimeCreated:     cfg.TimeCreated,
		TimeUpdated:     cfg.TimeUpdated,
	}, nil
}

func (c *ConfigureStoreImpl) UpdateConfiguration(ctx context.Context, cfg *domains.Configure) error {
	configure := &Configure{
		ConfigId:        cfg.ConfigId,
		ConfigKey:       cfg.ConfigKey,
		ConfigNamespace: cfg.ConfigNamespace,
		ConfigEnv:       cfg.ConfigEnv,
		ConfigDc:        cfg.ConfigDc,
		ContentType:     cfg.ContentType,
		Content:         cfg.Content,
		ConfigStatus:    cfg.ConfigStatus,
		ConfigVersion:   cfg.ConfigVersion,
		TimeCreated:     cfg.TimeCreated,
		TimeUpdated:     cfg.TimeUpdated,
	}
	sess := dbtxn.GetTxn(ctx)
	if _, err := sess.ID(configure.ConfigId).Update(configure); err != nil {
		return err
	}
	return nil
}

func (c *ConfigureStoreImpl) NextConfigVersionSeq(ctx context.Context) (int64, error) {
	sess := dbtxn.GetTxn(ctx)
	result, err := sess.QueryString(`select nextval('onlyconfig_version_seq') as seq`)
	if err != nil {
		return 0, err
	}
	val := result[0]["seq"]
	nextVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return nextVal, nil
}
