package main

import (
	"encoding/json"
	"io"
	"net/http"

	"google.golang.org/api/run/v1"
)

type ColumnDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	Description string `json:Description`
	Priority    int    `json:Priority`
}

type TableRow struct {
	Cells []interface{} `json:"cells"`
}

type TableResponse struct {
	Kind              string             `json:"kind"`
	APIVersion        string             `json:"apiVersion"`
	ColumnDefinitions []ColumnDefinition `json:"columnDefinitions"`
	Rows              []TableRow         `json:"rows"`
}

func tableHandler(w http.ResponseWriter, resp io.Reader) {
	w.Header().Set("content-type", "application/json")

	var listResp run.ListServicesResponse
	if err := json.NewDecoder(resp).Decode(&listResp); err != nil {
		panic(err) // TODO(ahmetb) don't panic
	}

	tableResp := TableResponse{
		Kind:       "Table",
		APIVersion: "meta.k8s.io/v1",
		ColumnDefinitions: []ColumnDefinition{
			{
				Name:        "Name",
				Type:        "string",
				Format:      "",
				Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
				Priority:    0,
			},
			{
				Name:        "URL",
				Type:        "string",
				Format:      "",
				Description: "Custom resource definition column (in JSONPath format): .status.url",
				Priority:    0,
			},
			{
				Name:        "LatestCreated",
				Type:        "string",
				Format:      "",
				Description: "Custom resource definition column (in JSONPath format): .status.latestCreatedRevisionName",
				Priority:    0,
			},
			{
				Name:        "LatestReady",
				Type:        "string",
				Format:      "",
				Description: "Custom resource definition column (in JSONPath format): .status.latestReadyRevisionName",
				Priority:    0,
			},
			{
				Name:        "Ready",
				Type:        "string",
				Format:      "",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].status",
				Priority:    0,
			},
			{
				Name:        "Reason",
				Type:        "string",
				Format:      "",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].reason",
				Priority:    0,
			},
		},
	}

	for _, item := range listResp.Items {
		row := TableRow{}
		row.Cells = append(row.Cells, item.Metadata.Name)
		row.Cells = append(row.Cells, item.Status.Url)
		row.Cells = append(row.Cells, item.Status.LatestCreatedRevisionName)
		row.Cells = append(row.Cells, item.Status.LatestReadyRevisionName)

		var readyCond *run.GoogleCloudRunV1Condition
		for _, c := range item.Status.Conditions {
			if c.Type == "Ready" {
				readyCond = c
				break
			}
		}
		if readyCond == nil {
			row.Cells = append(row.Cells, nil)
			row.Cells = append(row.Cells, nil)
		} else {
			row.Cells = append(row.Cells, readyCond.Status)
			row.Cells = append(row.Cells, readyCond.Reason)
		}
		tableResp.Rows = append(tableResp.Rows, row)
	}

	if err := json.NewEncoder(w).Encode(tableResp); err != nil {
		panic(err) // TODO(ahmetb) don't panic here
	}
}
