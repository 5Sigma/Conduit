package server

import (
	"conduit/log"
	"net/http"
	"postmaster/api"
	"postmaster/mailbox"
)

// deployInfo is used to both report a list of recent deployments and get
// details and responses on a specific deployment.
func deployInfo(w http.ResponseWriter, r *http.Request) {
	var request api.DeploymentStatsRequest
	err := readRequest(r, &request)
	if err != nil {
		sendError(w, "Could not parse request")
		return
	}

	if request.NamePattern == "" {
		request.NamePattern = ".*"
	}

	if request.TokenPattern == "" {
		request.TokenPattern = ".*"
	}

	accessKey, err := mailbox.FindKeyByName(request.AccessKeyName)
	if err != nil || accessKey == nil {
		sendError(w, "Access key is invalid")
		return
	}
	if !accessKey.CanAdmin() {
		sendError(w, "Not allowed to list deployments")
		return
	}
	if !request.Validate(accessKey.Secret) {
		sendError(w, "Signature invalid")
		return
	}
	response := api.DeploymentStatsResponse{}
	if request.Deployment == "" {
		log.Infof("Listing all deploys for %s", accessKey.Name)
		deployments, err := mailbox.ListDeployments(request.NamePattern,
			int(request.Count), request.TokenPattern)
		if err != nil {
			sendError(w, err.Error())
			return
		}
		for _, d := range deployments {
			dStats, err := d.Stats()
			if err != nil {
				sendError(w, err.Error())
				return
			}
			statsResp := api.DeploymentStats{
				Name:          d.Name,
				Id:            d.Id,
				PendingCount:  dStats.PendingCount,
				MessageCount:  dStats.MessageCount,
				ResponseCount: dStats.ResponseCount,
				CreatedAt:     d.DeployedAt,
				DeployedBy:    d.DeployedBy,
			}
			response.Deployments = append(response.Deployments, statsResp)
		}
	} else {
		dep, err := mailbox.FindDeployment(request.Deployment)
		if err != nil {
			sendError(w, err.Error())
			return
		}

		if dep == nil {
			sendError(w, "Deployment not found")
			return
		}

		dStats, err := dep.Stats()
		if err != nil {
			sendError(w, err.Error())
			return
		}
		deploymentResponses, err := dep.GetResponses()
		if err != nil {
			sendError(w, err.Error())
			return
		}
		statsResp := api.DeploymentStats{
			Name:          dep.Name,
			Id:            dep.Id,
			PendingCount:  dStats.PendingCount,
			MessageCount:  dStats.MessageCount,
			ResponseCount: dStats.ResponseCount,
			CreatedAt:     dep.DeployedAt,
			Responses:     []api.DeploymentResponse{},
		}
		for _, r := range deploymentResponses {
			apiR := api.DeploymentResponse{
				Mailbox:     r.Mailbox,
				Response:    r.Response,
				RespondedAt: r.RespondedAt,
				IsError:     r.IsError,
			}
			statsResp.Responses = append(statsResp.Responses, apiR)
		}
		response.Deployments = append(response.Deployments, statsResp)
	}
	response.Sign(accessKey.Name, accessKey.Secret)
	writeResponse(&w, response)
}
