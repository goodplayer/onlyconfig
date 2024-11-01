package controller

import (
	"context"
	"log"
	"net/http"
	"strings"

	"gitea.com/go-chi/binding"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"xorm.io/xorm"

	"github.com/goodplayer/onlyconfig/webmgr/config"
	"github.com/goodplayer/onlyconfig/webmgr/domains"
	"github.com/goodplayer/onlyconfig/webmgr/storage/postgres"
)

const (
	JwtTokenContextKey = "jwt_token"
)

func (cc *ControllerContainer) addUserControllers(r *chi.Mux, engine *xorm.Engine) {

	userStore := &postgres.UserStoreImpl{}
	u := &UserController{
		TxnController: TxnController{
			Engine: engine,
		},
		UserHandler: &domains.UserHandler{
			UserStore: userStore,
			Config:    cc.CfgVal,
		},
	}
	cc.UserHandler = u.UserHandler

	r.Post("/auth/user/login", u.UserLogin)
	r.Post("/user/new_user", u.UserRegister)

	r.Group(func(r chi.Router) {
		r.Use(cc.userJwtTokenMiddleware())
		r.Get("/user/organizations", u.QueryUserOrganizations)
		r.Post("/user/change_password", u.ChangePassword)
		r.Put("/organization/{org_name}", u.CreateOrganization)
		r.Put("/organization/{org_name}/owner/{username}", u.AddOwnerToOrg)
	})
}

func (cc *ControllerContainer) userJwtTokenMiddleware() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
				bearer = strings.TrimSpace(bearer[7:])
			} else {
				Unauthorized(w, r)
				return
			}
			// jwt validation
			if ok, err := cc.UserHandler.ValidateUserJwtToken(bearer); err != nil {
				log.Println("validate jwt token failed:", err)
				Unauthorized(w, r)
				return
			} else if !ok {
				log.Println("validate jwt token failed.")
				Unauthorized(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), JwtTokenContextKey, bearer)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

type UserController struct {
	TxnController
	UserHandler *domains.UserHandler
}

func (u *UserController) UserLogin(w http.ResponseWriter, r *http.Request) {
	loginEntity := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := binding.JSON(r, &loginEntity); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if loginEntity.Username == "" || loginEntity.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		user, err := u.UserHandler.LoginUser(ctx, loginEntity.Username, loginEntity.Password)
		if err != nil {
			log.Println("error login user:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusUnauthorized)
			}, TxnStatusRollback
		}

		token, uuid, err := user.GenerateJwtToken(u.UserHandler.Config.Load().(*config.WebManagerConfig).JwtSecrets[0])
		if err != nil {
			log.Println("error:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}

		log.Println("login uuid:", uuid)

		return func(writer http.ResponseWriter, request *http.Request) {
			render.JSON(writer, request, map[string]interface{}{
				"token": token,
			})
		}, TxnStatusCommit
	})
}

func (u *UserController) QueryUserOrganizations(w http.ResponseWriter, r *http.Request) {
	u.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		jwtToken := r.Context().Value(JwtTokenContextKey).(string)
		claims, err := u.UserHandler.GetClaimsFromJwtToken(jwtToken)
		if err != nil {
			log.Println("get claims failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusUnauthorized)
			}, TxnStatusRollback
		}
		orgs, err := u.UserHandler.QueryOrganizationsByUserId(ctx, claims.UserId)
		if err != nil {
			log.Println("query organizations failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		var result []map[string]interface{}
		for _, org := range orgs {
			var userList, ownerList []string
			for _, user := range org.UserList {
				userList = append(userList, user.UserName)
			}
			for _, owner := range org.OwnerList {
				ownerList = append(ownerList, owner.UserName)
			}
			result = append(result, map[string]interface{}{
				"org_id":     org.OrgId,
				"org_name":   org.OrgName,
				"owner_list": ownerList,
				"user_list":  userList,
			})
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			render.JSON(writer, request, map[string]any{
				"result": result,
			})
		}, TxnStatusCommit
	})
}

func (u *UserController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	u.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		orgName := strings.TrimSpace(chi.URLParam(r, "org_name"))
		if orgName == "" {
			log.Println("empty org name")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		jwtToken := r.Context().Value(JwtTokenContextKey).(string)
		claims, err := u.UserHandler.GetClaimsFromJwtToken(jwtToken)
		if err != nil {
			log.Println("get claims failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusUnauthorized)
			}, TxnStatusRollback
		}
		if err := u.UserHandler.CreateOrganization(ctx, orgName, claims.Username); err != nil {
			log.Println("create organization failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (u *UserController) AddOwnerToOrg(w http.ResponseWriter, r *http.Request) {
	u.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		orgName := strings.TrimSpace(chi.URLParam(r, "org_name"))
		username := strings.TrimSpace(chi.URLParam(r, "username"))
		if orgName == "" || username == "" {
			log.Println("empty org_name or username")
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if err := u.UserHandler.AddUserToOrg(ctx, username, orgName, 1); err != nil {
			log.Println("add user to org error:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (u *UserController) UserRegister(w http.ResponseWriter, r *http.Request) {
	u.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		req := new(domains.UserRegisterReq)
		if err := render.DefaultDecoder(r, req); err != nil {
			log.Println("request body failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		if err := u.UserHandler.UserRegister(ctx, req); err != nil {
			log.Println("register user failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}

func (u *UserController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u.RunInTxn(w, r, func(ctx context.Context) (RenderFn, TxnStatus) {
		jwtToken := r.Context().Value(JwtTokenContextKey).(string)
		claims, err := u.UserHandler.GetClaimsFromJwtToken(jwtToken)
		if err != nil {
			log.Println("get claims failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusUnauthorized)
			}, TxnStatusRollback
		}
		req := &domains.ChangePasswordReq{}
		if err := render.DefaultDecoder(r, req); err != nil {
			log.Println("request body failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusBadRequest)
			}, TxnStatusRollback
		}
		req.Username = claims.Username
		if err := u.UserHandler.ChangePassword(ctx, req); err != nil {
			log.Println("change password failed:", err)
			return func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusInternalServerError)
			}, TxnStatusRollback
		}
		return func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}, TxnStatusCommit
	})
}
