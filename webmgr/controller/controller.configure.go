package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"xorm.io/xorm"

	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/postgres"
	"github.com/goodplayer/onlyconfig/webmgr/tools"
)

func (cc *ControllerContainer) addConfigureControllers(r *chi.Mux, engine *xorm.Engine) {
	ccl := &ConfigureController{
		TxnController: TxnController{
			Engine: engine,
		},
		ConfigureHandler: &domains.ConfigureHandler{
			ConfigureRepository:  &postgres.ConfigureStoreImpl{},
			PushChangeRepository: &postgres.PushChangeRepositoryImpl{},
		},
		UserHandler: cc.UserHandler,
	}

	r.Route("/configures", func(r chi.Router) {
		r.Use(cc.userJwtTokenMiddleware())

		r.Get("/applications", ccl.Applications)
		r.Put("/application/{org_id}/{app_name}", ccl.CreateApplication)
		r.Put("/application/{app_id}/{env}/{dc}", ccl.AddAppEnvAndDc)
		r.Get("/env_dc_list", ccl.EnvDcList)
		r.Put("/env_and_dc/{type}/{name}", ccl.AddEnvAndDc)
		r.Put("/namespace/{app_id}/{ns_name}/{ns_type}", ccl.AddAppNamespace)
		r.Get("/namespaces/{app_id}", ccl.QueryAppNamespaces)
		r.Get("/configure_list/{app_id}/{env}/{dc}", ccl.QueryAppConfigList)
		r.Post("/configure/{app_id}/{env}/{dc}/{namespace}/{key}", ccl.AddConfiguration)
		r.Get("/configure/{cfg_id}", ccl.QueryConfigById)
		r.Put("/configure/{cfg_id}", ccl.UpdateConfigById)
	})
}

type ConfigureController struct {
	TxnController

	ConfigureHandler *domains.ConfigureHandler
	UserHandler      *domains.UserHandler
}

func (c *ConfigureController) EnvDcList(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		var dcList []string
		var envList []string
		// query env and dc list from database
		if dcs, envs, err := c.ConfigureHandler.LoadEnvAndDcList(ctx); err != nil {
			log.Println("load env and dc list failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		} else {
			for _, dc := range dcs {
				dcList = append(dcList, dc.DatacenterName)
			}
			for _, env := range envs {
				envList = append(envList, env.EnvName)
			}
		}

		return func(writer http.ResponseWriter, request *http.Request) {
			result := map[string]any{
				"env": envList,
				"dc":  dcList,
			}
			render.JSON(writer, request, result)
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) AddEnvAndDc(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		addType := strings.TrimSpace(chi.URLParam(r, "type"))
		addName := strings.TrimSpace(chi.URLParam(r, "name"))
		if addType != "dc" && addType != "env" {
			log.Println("unknown addType:", addType)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if addName == "" {
			log.Println("empty addName")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if !tools.ValidateName(addName) {
			log.Println("invalid format of addName:", addName)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}

		// add env and dc
		if err := c.ConfigureHandler.AddEnvDc(ctx, addType, addName); errors.Is(err, domains.ErrDuplicatedEnvOrDc) {
			log.Println("duplicated env or dc:", addType, addName)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusConflict)
			}, TxnStatusRollback
		} else if err != nil {
			log.Println("add env or dc failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}

		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) Applications(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		jwtToken := r.Context().Value(JwtTokenContextKey).(string)
		claims, err := c.UserHandler.GetClaimsFromJwtToken(jwtToken)
		if err != nil {
			log.Println("get claims failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusUnauthorized)
			}, TxnStatusRollback
		}
		orgs, err := c.UserHandler.QueryOrganizationsByUserId(ctx, claims.UserId)
		if err != nil {
			log.Println("query organizations failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		applicationList, err := c.ConfigureHandler.LoadApplicationsByOrganizationIdList(ctx, orgs)
		if err != nil {
			log.Println("load applications failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		var result []map[string]any
		for _, application := range applicationList {
			var item []map[string]any
			for _, elem := range application.EnvAndDcList {
				item = append(item, map[string]any{
					"env":     elem.EnvName,
					"dc_list": elem.DcList,
				})
			}
			result = append(result, map[string]any{
				"app_id":             application.ApplicationId,
				"app_name":           application.ApplicationName,
				"app_desc":           application.ApplicationDescription,
				"app_owner_org_id":   application.ApplicationOwnerOrganization.OrgId,
				"app_owner_org_name": application.ApplicationOwnerOrganization.OrgName,
				"env_and_dc":         item,
			})
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			render.JSON(writer, request, map[string]any{
				"list": result,
			})
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) CreateApplication(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		orgId := strings.TrimSpace(chi.URLParam(r, "org_id"))
		appName := strings.TrimSpace(chi.URLParam(r, "app_name"))
		if orgId == "" || appName == "" {
			log.Println("empty orgId or appName")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if !tools.ValidateName(appName) {
			log.Println("invalid format of appName:", appName)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}

		org, err := c.UserHandler.LoadOrganizationByOrgId(ctx, orgId)
		if err != nil {
			log.Println("load organization failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}

		if err := c.ConfigureHandler.CreateApplication(ctx, org, appName); err != nil {
			log.Println("create application failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}

		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) AddAppEnvAndDc(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		appIdStr := strings.TrimSpace(chi.URLParam(r, "app_id"))
		env := strings.TrimSpace(chi.URLParam(r, "env"))
		dc := strings.TrimSpace(chi.URLParam(r, "dc"))
		if appIdStr == "" || env == "" || dc == "" {
			log.Println("empty appId or env or dc")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		appId, err := strconv.ParseInt(appIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of appId:", appIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if err := c.ConfigureHandler.LinkEnvAndDcToApp(ctx, env, dc, appId); err != nil {
			log.Println("link env and dc to app failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) AddAppNamespace(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		appIdStr := strings.TrimSpace(chi.URLParam(r, "app_id"))
		nsName := strings.TrimSpace(chi.URLParam(r, "ns_name"))
		nsType := strings.TrimSpace(chi.URLParam(r, "ns_type"))
		if appIdStr == "" || nsName == "" || nsType == "" {
			log.Println("empty appId or nsName or nsType")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		appId, err := strconv.ParseInt(appIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of appId:", appIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if !tools.ValidateName(nsName) {
			log.Println("invalid format of nsName:", nsName)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		switch nsType {
		case "public":
		case "application":
		default:
			log.Println("invalid format of nsType:", nsType)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}

		if err := c.ConfigureHandler.AddApplicationNamespace(ctx, appId, nsName, nsType); err != nil {
			log.Println("add namespace failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) QueryAppNamespaces(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		appIdStr := strings.TrimSpace(chi.URLParam(r, "app_id"))
		appId, err := strconv.ParseInt(appIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of appId:", appIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		nsList, err := c.ConfigureHandler.QueryApplicationNamespaces(ctx, appId)
		if err != nil {
			log.Println("query namespace failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		var result []string
		for _, ns := range nsList {
			result = append(result, ns.Name)
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			render.JSON(writer, request, map[string]any{
				"result": result,
			})
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) AddConfiguration(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		appIdStr := strings.TrimSpace(chi.URLParam(r, "app_id"))
		env := strings.TrimSpace(chi.URLParam(r, "env"))
		dc := strings.TrimSpace(chi.URLParam(r, "dc"))
		namespace := strings.TrimSpace(chi.URLParam(r, "namespace"))
		key := strings.TrimSpace(chi.URLParam(r, "key"))
		if appIdStr == "" || env == "" || dc == "" || namespace == "" || key == "" {
			log.Println("empty appId or env or dc or namespace or key")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		appId, err := strconv.ParseInt(appIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of appId:", appIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		req := &domains.AddConfigurationRequest{
			AppId:     appId,
			Env:       env,
			Dc:        dc,
			Namespace: namespace,
			Key:       key,
		}
		if err := render.DefaultDecoder(r, req); err != nil {
			log.Println("bind failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		switch req.ContentType {
		case "general":
		default:
			log.Println("invalid format of content type:", req.ContentType)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}

		if err := c.ConfigureHandler.AddConfiguration(ctx, req); err != nil {
			log.Println("add configuration failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) QueryAppConfigList(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		appIdStr := strings.TrimSpace(chi.URLParam(r, "app_id"))
		env := strings.TrimSpace(chi.URLParam(r, "env"))
		dc := strings.TrimSpace(chi.URLParam(r, "dc"))
		if appIdStr == "" || env == "" || dc == "" {
			log.Println("empty appId or env or dc")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		appId, err := strconv.ParseInt(appIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of appId:", appIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		configResult, err := c.ConfigureHandler.QueryAppConfigList(ctx, appId, env, dc)
		if err != nil {
			log.Println("query configuration failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		var result []struct {
			Namespace  string `json:"namespace"`
			ConfigList []struct {
				Key      string `json:"key"`
				ConfigId string `json:"configure_id"`
			} `json:"configure_list"`
		}
		for key, configs := range configResult {
			var cfgList []struct {
				Key      string `json:"key"`
				ConfigId string `json:"configure_id"`
			}
			for _, config := range configs {
				cfgList = append(cfgList, struct {
					Key      string `json:"key"`
					ConfigId string `json:"configure_id"`
				}{Key: config.ConfigKey, ConfigId: fmt.Sprint(config.ConfigId)})
			}
			slices.SortFunc(cfgList, func(a, b struct {
				Key      string `json:"key"`
				ConfigId string `json:"configure_id"`
			}) int {
				return strings.Compare(a.Key, b.Key)
			})
			result = append(result, struct {
				Namespace  string `json:"namespace"`
				ConfigList []struct {
					Key      string `json:"key"`
					ConfigId string `json:"configure_id"`
				} `json:"configure_list"`
			}{Namespace: key, ConfigList: cfgList})
		}
		slices.SortFunc(result, func(a, b struct {
			Namespace  string `json:"namespace"`
			ConfigList []struct {
				Key      string `json:"key"`
				ConfigId string `json:"configure_id"`
			} `json:"configure_list"`
		}) int {
			return strings.Compare(a.Namespace, b.Namespace)
		})
		return func(writer http.ResponseWriter, request *http.Request) {
			render.JSON(writer, request, map[string]any{
				"result": result,
			})
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) QueryConfigById(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		cfgIdStr := strings.TrimSpace(chi.URLParam(r, "cfg_id"))
		cfgId, err := strconv.ParseInt(cfgIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of cfgId:", cfgIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}

		cfg, err := c.ConfigureHandler.QueryConfigureById(ctx, cfgId)
		if err != nil {
			log.Println("query configuration failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			render.JSON(writer, request, map[string]any{
				"result": map[string]any{
					"cfg_id":      fmt.Sprint(cfg.ConfigId),
					"cfg_key":     cfg.ConfigKey,
					"cfg_ns":      cfg.ConfigNamespace,
					"cfg_env":     cfg.ConfigEnv,
					"cfg_dc":      cfg.ConfigDc,
					"cfg_status":  cfg.ConfigStatus,
					"cfg_ct":      cfg.ContentType,
					"cfg_content": cfg.Content,
				},
			})
		}, TxnStatusCommit
	})
}

func (c *ConfigureController) UpdateConfigById(w http.ResponseWriter, r *http.Request) {
	c.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		cfgIdStr := strings.TrimSpace(chi.URLParam(r, "cfg_id"))
		cfgId, err := strconv.ParseInt(cfgIdStr, 10, 64)
		if err != nil {
			log.Println("invalid format of cfgId:", cfgIdStr)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}

		req := &domains.UpdateConfigurationRequest{
			ConfigId: cfgId,
		}
		if err := render.DefaultDecoder(r, req); err != nil {
			log.Println("bind failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if err := c.ConfigureHandler.UpdateConfigurationById(ctx, req); err != nil {
			log.Println("update configuration failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}
