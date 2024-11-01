package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/meidoworks/nekoq-component/configure/configapi"

	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/dbtxn"
)

type Configuration struct {
	ConfigId          int64  `xorm:"'cfg_id' pk autoincr"`
	Selectors         string `xorm:"'selectors'"`
	OptionalSelectors string `xorm:"'optional_selectors' default ''"`
	ConfigGroup       string `xorm:"'cfg_group'"`
	ConfigKey         string `xorm:"'cfg_key'"`
	ConfigVersion     string `xorm:"'cfg_version'"`
	RawConfigValue    []byte `xorm:"'raw_cfg_value'"`
	ConfigStatus      int64  `xorm:"'cfg_status'"`
	TimeCreated       int64  `xorm:"'time_created'"`
	TimeUpdated       int64  `xorm:"'time_updated'"`
	Sequence          int64  `xorm:"'sequence'"`
}

func (c *Configuration) TableName() string {
	return "configuration"
}

type PushChangeRepositoryImpl struct{}

func (p *PushChangeRepositoryImpl) getSelectorsString(cfg *domains.Configure, app *domains.Application) string {
	return cfg.GenerateSelectorsString(app)
}

func (p *PushChangeRepositoryImpl) getOptSelectorsString(cfg *domains.Configure, app *domains.Application) string {
	return ""
}

func (p *PushChangeRepositoryImpl) getSelectors(cfg *domains.Configure, app *domains.Application) configapi.Selectors {
	return *cfg.GenerateSelectors(app)
}

func (p *PushChangeRepositoryImpl) getOptSelectors(cfg *domains.Configure, app *domains.Application) configapi.Selectors {
	return configapi.Selectors{}
}

func (p *PushChangeRepositoryImpl) ExistsConfiguration(ctx context.Context, cfg *domains.Configure, app *domains.Application) (bool, error) {
	configuration := new(Configuration)
	sess := dbtxn.GetTxn(ctx)
	if has, err := sess.Where("selectors = ? and optional_selectors = ? and cfg_group = ? and cfg_key = ?", p.getSelectorsString(cfg, app), p.getOptSelectorsString(cfg, app), cfg.ConfigNamespace, cfg.ConfigKey).Get(configuration); err != nil {
		return false, err
	} else {
		return has, nil
	}
}

func (p *PushChangeRepositoryImpl) InsertNewConfigure(ctx context.Context, cfg *domains.Configure, app *domains.Application) (int64, error) {
	sess := dbtxn.GetTxn(ctx)

	now := time.Now()
	sig := sha256.Sum256([]byte(cfg.Content))
	apiCfg := &configapi.Configuration{
		Group:             cfg.ConfigNamespace,
		Key:               cfg.ConfigKey,
		Version:           cfg.ConfigVersion,
		Value:             []byte(cfg.Content),
		Signature:         "sha256:" + hex.EncodeToString(sig[:]),
		Selectors:         p.getSelectors(cfg, app),
		OptionalSelectors: p.getOptSelectors(cfg, app),
		Timestamp:         now.Unix(),
	}
	data, err := cbor.Marshal(apiCfg)
	if err != nil {
		return 0, err
	}
	configuration := &Configuration{
		Selectors:         p.getSelectorsString(cfg, app),
		OptionalSelectors: p.getOptSelectorsString(cfg, app),
		ConfigGroup:       cfg.ConfigNamespace,
		ConfigKey:         cfg.ConfigKey,
		ConfigVersion:     cfg.ConfigVersion,
		RawConfigValue:    data,
		ConfigStatus:      0,
		TimeCreated:       now.UnixMilli(),
		TimeUpdated:       now.UnixMilli(),
		Sequence:          0,
	}
	if _, err := sess.Insert(configuration); err != nil {
		return 0, err
	}

	return configuration.ConfigId, nil
}

func (p *PushChangeRepositoryImpl) UpdateConfigure(ctx context.Context, cfg *domains.Configure, app *domains.Application) (bool, error) {
	sess := dbtxn.GetTxn(ctx)

	configuration := new(Configuration)
	if has, err := sess.Where("selectors = ? and optional_selectors = ? and cfg_group = ? and cfg_key = ?", p.getSelectorsString(cfg, app), p.getOptSelectorsString(cfg, app), cfg.ConfigNamespace, cfg.ConfigKey).Get(configuration); err != nil {
		return false, err
	} else if !has {
		return false, errors.New("configuration not found")
	}

	now := time.Now()
	sig := sha256.Sum256([]byte(cfg.Content))
	apiCfg := &configapi.Configuration{
		Group:             cfg.ConfigNamespace,
		Key:               cfg.ConfigKey,
		Version:           cfg.ConfigVersion,
		Value:             []byte(cfg.Content),
		Signature:         "sha256:" + hex.EncodeToString(sig[:]),
		Selectors:         p.getSelectors(cfg, app),
		OptionalSelectors: p.getOptSelectors(cfg, app),
		Timestamp:         now.Unix(),
	}
	data, err := cbor.Marshal(apiCfg)
	if err != nil {
		return false, err
	}
	if r, err := sess.Exec("update configuration set cfg_version = ?, raw_cfg_value = ?, time_updated = ?, sequence = nextval('cfg_seq') where cfg_id = ? and pg_try_advisory_xact_lock(-1000)", cfg.ConfigVersion, data, now.UnixMilli(), configuration.ConfigId); err != nil {
		return false, err
	} else {
		if rowcnt, err := r.RowsAffected(); err != nil {
			return false, err
		} else if rowcnt == 0 {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func (p *PushChangeRepositoryImpl) UpdateConfigurationSequence(ctx context.Context, cfg *domains.Configure, app *domains.Application, configId int64) (bool, error) {
	sess := dbtxn.GetTxn(ctx)
	now := time.Now()
	if r, err := sess.Exec("update configuration set time_updated = ?, sequence = nextval('cfg_seq') where cfg_id = ? and pg_try_advisory_xact_lock(-1000)", now.UnixMilli(), configId); err != nil {
		return false, err
	} else {
		if rowcnt, err := r.RowsAffected(); err != nil {
			return false, err
		} else if rowcnt == 0 {
			return false, nil
		} else {
			return true, nil
		}
	}
}
