package cmd

import (
	"fmt"
	"github.com/5sigma/conduit/log"
	"github.com/5sigma/conduit/postmaster/api"
	"github.com/spf13/cobra"
)

type (
	consolidatedResponse struct {
		Response  string
		Mailboxes []string
		IsError   bool
	}

	consolidatedResponses []consolidatedResponse
)

// consolidateResponses will consolidate a list of DeploymentResponses into a
// single item per response. This is used when displaying responses to condense
// the list.
func consolidateResponses(
	responses api.DeploymentResponses,
) consolidatedResponses {
	cResponses := consolidatedResponses{}
	for _, r := range responses {
		index, exists := cResponses.HasResponse(r)
		if !exists {
			cr := consolidatedResponse{
				Mailboxes: []string{r.Mailbox},
				Response:  r.Response,
				IsError:   r.IsError,
			}
			cResponses = append(cResponses, cr)
		} else {
			cResponses[index].Mailboxes = append(cResponses[index].Mailboxes,
				r.Mailbox)
		}
	}
	return cResponses
}

// HasResponse is a helper function that returns true if the collection has the
// passed response. It also returns the index of the item in the list, or -1 if
// not present.
func (crs consolidatedResponses) HasResponse(r api.DeploymentResponse) (int, bool) {
	for i, cr := range crs {
		if cr.Response == r.Response && cr.IsError == r.IsError {
			return i, true
		}
	}
	return -1, false
}

// A helper function used by both the 'deploy' and 'deploy get' commands to
// display a list of responses for a deployment.
func displayResponses(rs api.DeploymentResponses, consolidate bool) {
	if consolidate {
		cResponses := consolidateResponses(rs)
		for _, r := range cResponses {
			var mbStr string
			if len(r.Mailboxes) == 1 {
				mbStr = r.Mailboxes[0]
			} else {
				mbStr = fmt.Sprintf("%d/%d", len(r.Mailboxes), len(rs))
			}
			if r.IsError == true {
				log.Errorf("%s: %s", mbStr, r.Response)
			} else {
				log.Infof("%s: %s", mbStr, r.Response)
			}
		}
	} else {
		for _, r := range rs {
			if r.IsError == true {
				log.Errorf("%s: %s", r.Mailbox, r.Response)
			} else {
				log.Infof("%s: %s", r.Mailbox, r.Response)
			}
		}
	}
}

// getDeploymentInfo returns the information about a deployment.
func getDeploymentInfo(depId string, consolidate bool) {
	client, _ := AdminClientFromConfig()
	stats, err := client.PollDeployment(depId,
		func(stats *api.DeploymentStats) bool {
			return false
		})
	if err != nil {
		log.Debug(err.Error())
		log.Error("Could not get deployment results")
		return
	}
	log.Info("")
	log.Alert(depId)
	log.Info("")
	log.Infof("Total messages: %d", stats.MessageCount)
	log.Infof("Pending messages: %d", stats.PendingCount)
	log.Infof("Total responses: %d", stats.ResponseCount)
	if len(stats.Responses) > 0 {
		log.Alert("\nResponses:")
	}
	stats.Responses.Sort()
	displayResponses(stats.Responses, consolidate)
	log.Info("")
}

// deployGetCmd represents the deployGet command
var deployGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get deployment information",
	Long: `Get deployment responses and statistics. You must specify the
deployment name. This name can be set when the deployment is made using the
"--name" flag. If a name was not specified a random name was assigned to it.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		} else {
			getDeploymentInfo(args[0], cmd.Flag("expand").Value.String() != "true")
		}
	},
}

func init() {
	deployCmd.AddCommand(deployGetCmd)
	deployGetCmd.Flags().BoolP("expand", "e", false,
		"Expand results (dont consolidate)")
}
