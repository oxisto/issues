// Copyright 2019 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package issues

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v29/github"
)

type RepositoryRefArray []int64

type Workspace struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	RepositoryIDs RepositoryRefArray `json:"repositoryIDs" gorm:"type:integer[]"`
}

type Relationship struct {
	IssueID      int64  `josn:"issueId" gorm:"primary_key;auto_increment:false"`
	OtherIssueID int64  `json:"otherIssueId" gorm:"primary_key;auto_increment:false"`
	Type         string `json:"type"`
}

func (r *RepositoryRefArray) Scan(src interface{}) error {
	if src == nil {
		*r = nil
		return nil
	}

	u, ok := src.([]uint8)
	if !ok {
		return errors.New("Unable to convert type from []uint8")
	}

	var intArray []int64
	var i int64
	var err error
	var s string

	s = strings.ReplaceAll(strings.ReplaceAll(string(u), "{", ""), "}", "")

	// split array
	array := strings.Split(s, ",")
	for _, v := range array {
		if i, err = strconv.ParseInt(v, 10, 64); err != nil {
			return fmt.Errorf("Could not convert all array elements to int64: %s", err)
		}

		intArray = append(intArray, i)
	}

	*r = intArray

	return nil
}

func (app *Application) GetWorkspace(workspaceID int64) (*Workspace, error) {
	return app.db.GetWorkspace(workspaceID)
}

type issue struct {
	Number int
	Title  string
}

// GetBacklog retrieves all issues from all workspaces that do not have a milestone
// associated with it. This needs to query the database as well as the GitHub
// GraphQL endpoint
func (app *Application) GetBacklog(clients *GitHubClients, workspaceID int64) (interface{}, error) {
	// just some fun for now
	/*var q struct {
		Repository struct {
			Issue struct {
				Nodes []issue
			} `graphql:"issues(last: 100, filterBy: {milestone: null, states: OPEN})"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String("oxisto"),
		"repositoryName":  githubv4.String("aybaze"),
	}

	err := clients.V4.Query(context.Background(), &q, variables)
	if err != nil {
		return nil
	}

	return q.Repository.Issue.Nodes*/
	start := time.Now()
	options := github.IssueListByRepoOptions{
		Milestone: "none",
		State:     "open",
	}

	issues, _, err := clients.V3.Issues.ListByRepo(context.Background(), "oxisto", "aybaze", &options)

	end := time.Now()

	duration := end.Sub(start)

	log.Infof("call to GitHub took %+v", duration)

	return issues, err
}
