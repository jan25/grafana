/*Package api contains base API implementation of unified alerting
 *
 *Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 *
 *Do not manually edit these files, please find ngalert/api/swagger-codegen/ for commands on how to generate them.
 */
package api

import (
	"github.com/go-macaron/binding"

	apimodels "github.com/grafana/alerting-api/pkg/api"
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/middleware"
	"github.com/grafana/grafana/pkg/models"
)

type AlertmanagerApiService interface {
	RouteCreateSilence(*models.ReqContext, apimodels.PostableSilence) response.Response
	RouteDeleteAlertingConfig(*models.ReqContext) response.Response
	RouteDeleteSilence(*models.ReqContext) response.Response
	RouteGetAMAlertGroups(*models.ReqContext) response.Response
	RouteGetAMAlerts(*models.ReqContext) response.Response
	RouteGetAlertingConfig(*models.ReqContext) response.Response
	RouteGetSilence(*models.ReqContext) response.Response
	RouteGetSilences(*models.ReqContext) response.Response
	RoutePostAMAlerts(*models.ReqContext, apimodels.PostableAlerts) response.Response
	RoutePostAlertingConfig(*models.ReqContext, apimodels.PostableUserConfig) response.Response
}

func (api *API) RegisterAlertmanagerApiEndpoints(srv AlertmanagerApiService) {
	api.RouteRegister.Group("", func(group routing.RouteRegister) {
		group.Post(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/silences"), binding.Bind(apimodels.PostableSilence{}), routing.Wrap(srv.RouteCreateSilence))
		group.Delete(toMacaronPath("/api/alertmanager/{Recipient}/config/api/v1/alerts"), routing.Wrap(srv.RouteDeleteAlertingConfig))
		group.Delete(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/silence/{SilenceId}"), routing.Wrap(srv.RouteDeleteSilence))
		group.Get(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/alerts/groups"), routing.Wrap(srv.RouteGetAMAlertGroups))
		group.Get(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/alerts"), routing.Wrap(srv.RouteGetAMAlerts))
		group.Get(toMacaronPath("/api/alertmanager/{Recipient}/config/api/v1/alerts"), routing.Wrap(srv.RouteGetAlertingConfig))
		group.Get(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/silence/{SilenceId}"), routing.Wrap(srv.RouteGetSilence))
		group.Get(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/silences"), routing.Wrap(srv.RouteGetSilences))
		group.Post(toMacaronPath("/api/alertmanager/{Recipient}/api/v2/alerts"), binding.Bind(apimodels.PostableAlerts{}), routing.Wrap(srv.RoutePostAMAlerts))
		group.Post(toMacaronPath("/api/alertmanager/{Recipient}/config/api/v1/alerts"), binding.Bind(apimodels.PostableUserConfig{}), routing.Wrap(srv.RoutePostAlertingConfig))
	}, middleware.ReqSignedIn)
}