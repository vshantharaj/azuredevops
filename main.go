package main

import (
	"context"
	"log"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	wit "github.com/microsoft/azure-devops-go-api/azuredevops/workitemtracking"
)

func main() {
	organizationUrl := "<url>"
	personalAccessToken := "<pattoken>"

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)

	ctx := context.Background()

	// Create a client to interact with the Core area
	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}
	branchName := "master"
	searchType := git.GitVersionType("branch")

	gitvesioncriteria := git.GitVersionDescriptor{
		Version:     &branchName,
		VersionType: &searchType,
	}

	includeworkitems := true
	searchlimit := 20
	gitquerycommetcriteria := git.GitQueryCommitsCriteria{
		Top:              &searchlimit,
		IncludeWorkItems: &includeworkitems,
		ItemVersion:      &gitvesioncriteria,
	}

	projectName := "es-TLM-federation"
	repoName := "es-TLM-Federation2"
	// Get first page of the list of team projects for your organization
	responseValue, err := gitClient.GetCommitsBatch(ctx, git.GetCommitsBatchArgs{
		SearchCriteria: &gitquerycommetcriteria,
		Project:        &projectName,
		RepositoryId:   &repoName,
	})
	if err != nil {
		log.Fatal(err)
	}

	index := 0
	uniqueworkitems := make(map[string]struct{})
	if responseValue != nil {
		// Log the page of team project names
		for _, commit := range *responseValue {

			log.Printf("Name[%v] = %v", index, *commit.Comment)
			for i, workitem := range *commit.WorkItems {
				log.Printf("workitem[%v]= %v , %v", i, *workitem.Id, *workitem.Url)
				uniqueworkitems[*workitem.Id] = struct{}{}
			}

			index++
		}

		witClient, err := wit.NewClient(ctx, connection)
		if err != nil {
			log.Fatal(err)
		}
		workitemids := []int{}
		for k := range uniqueworkitems {
			val, _ := strconv.ParseInt(k, 10, 32)
			workitemids = append(workitemids, int(val))
		}
		workitembatchreq := wit.WorkItemBatchGetRequest{
			Ids: &workitemids,
		}

		wilist, err := witClient.GetWorkItemsBatch(ctx, wit.GetWorkItemsBatchArgs{
			WorkItemGetRequest: &workitembatchreq,
			Project:            &projectName,
		})
		if err != nil {
			log.Fatal(err)
		}

		for i, wi := range *wilist {
			if (*wi.Fields)["System.WorkItemType"] == "Product Backlog Item" {
				log.Printf("workitem[%v]= %v-%v\n%v\n%v", i, *wi.Id, (*wi.Fields)["System.Title"], (*wi.Fields)["System.IterationPath"], (*wi.Fields)["System.Tags"])
			}

		}
		// if continuationToken has a value, then there is at least one more page of projects to get
		// if responseValue.ContinuationToken != "" {
		// 	// Get next page of team projects
		// 	projectArgs := git.GetCommitsBatchArgs{
		// 		SearchCriteria: ,
		// 	}
		// 	responseValue, err = coreClient.GetCommitsBatch(ctx, projectArgs)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// } else {
		// 	responseValue = nil
		// }
	}
}
