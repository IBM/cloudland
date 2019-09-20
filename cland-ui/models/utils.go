package models

import (
	"fmt"
	"net/http"
	"net/url"

	restModels "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/facebookgo/errgroup"
	"github.com/mohae/deepcopy"
)

var backendEndpoint *url.URL
var IdentityPath string

func init() {
	var err error
	rawBackendEndpoint := beego.AppConfig.String("dev::backendEndpoint")
	IdentityPath = beego.AppConfig.String("dev::identityPath")
	backendEndpoint, err = url.Parse(rawBackendEndpoint)
	if err != nil {
		panic(fmt.Sprintf("Fail to parse endpoint URL: %s, error: %s", rawBackendEndpoint, err.Error()))
	}
}

func Authenticate(identityEndpoint, username, password, org string) (token string, err error) {
	authenEndpoint, _ := url.Parse(identityEndpoint)
	authenEndpoint.Path = authenEndpoint.Path + routeInfo[`identityToken`]
	logs.Debug(`try to get token with auth endpoint`, authenEndpoint.String())
	if org == `` {
		org = `default`
	}
	tokenResqStruct := &restModels.PostIdentityV3AuthTokensParamsBody{
		Auth: &restModels.PostIdentityV3AuthTokensParamsBodyAuth{
			Identity: &restModels.PostIdentityV3AuthTokensParamsBodyAuthIdentity{
				Password: &restModels.PostIdentityV3AuthTokensParamsBodyAuthIdentityPassword{
					User: &restModels.PostIdentityV3AuthTokensParamsBodyAuthIdentityPasswordUser{
						Name:     username,
						Password: password,
					},
				},
			},
		},
	}
	tokensInstance := &restModels.Token{}
	resp, respBody, errs := NewResut(http.MethodPost, authenEndpoint.String(), tokenResqStruct).EndStruct(tokensInstance)
	if len(errs) != 0 {
		err = fmt.Errorf("fail to call identiy servier to get token with status code: %d, response body: %s", resp.StatusCode, string(respBody))
		logs.Error(err)
		return
	}
	token = resp.Header.Get(TOKEN_HEADER_KEY)
	if token == `` {
		err = fmt.Errorf(`fail to get token key`)
		return
	}
	return token, nil
}

func Identity() (*restModels.GetIdentityMultipleChoicesBody, error) {
	identityEndpoint := deepcopy.Copy(backendEndpoint).(*url.URL)
	identityEndpoint.Path = IdentityPath
	logs.Debug("try to get identity infor with endpoint:", identityEndpoint.String())
	identity := &restModels.GetIdentityMultipleChoicesBody{}
	resp, respBody, errs := NewResut(http.MethodGet, identityEndpoint.String(), nil).EndStruct(identity)
	if len(errs) != 0 {
		err := errgroup.NewMultiError(errs...)
		logs.Error(fmt.Errorf(`call identity service fail with error: %s`, err))
		return nil, err
	}
	if resp.StatusCode != http.StatusMultipleChoices {
		errMsg := fmt.Errorf("fail to call identity servier with status code: %d, response body: %s", resp.StatusCode, string(respBody))
		logs.Error(errMsg)
		return nil, errMsg
	}
	return identity, nil
}
