package main

import (
	"encoding/json"
	"io"

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

func revisionTableConvert(resp io.Reader) TableResponse {
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
				Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			},
			{
				Name:        "Config Name",
				Type:        "string",
				Description: `Custom resource definition column (in JSONPath format): .metadata.labels['serving\.knative\.dev/configuration']`,
			},
			{
				Name:        "Generation",
				Type:        "string",
				Description: `Custom resource definition column (in JSONPath format): .metadata.labels['serving\\.knative\\.dev/configurationGeneration']`,
			},
			{
				Name:        "Ready",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].status",
			},
			{
				Name:        "Reason",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].reason",
			},
		},
	}

	for _, item := range listResp.Items {
		row := TableRow{}
		row.Cells = append(row.Cells, item.Metadata.Name)
		row.Cells = append(row.Cells, item.Metadata.Labels["serving.knative.dev/configuration"])
		row.Cells = append(row.Cells, item.Metadata.Labels["serving.knative.dev/configurationGeneration"])
		s, r, m := condition(item.Status.Conditions, "Ready")
		row.Cells = append(row.Cells, s, r, m)
		tableResp.Rows = append(tableResp.Rows, row)
	}

	return tableResp
}

func ksvcTableConvert(resp io.Reader) TableResponse {
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
				Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			},
			{
				Name:        "URL",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.url",
			},
			{
				Name:        "LatestCreated",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.latestCreatedRevisionName",
			},
			{
				Name:        "LatestReady",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.latestReadyRevisionName",
			},
			{
				Name:        "Ready",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].status",
			},
			{
				Name:        "Reason",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].reason",
			},
		},
	}

	for _, item := range listResp.Items {
		row := TableRow{}
		row.Cells = append(row.Cells, item.Metadata.Name)
		row.Cells = append(row.Cells, item.Status.Url)
		row.Cells = append(row.Cells, item.Status.LatestCreatedRevisionName)
		row.Cells = append(row.Cells, item.Status.LatestReadyRevisionName)
		s, r, m := condition(item.Status.Conditions, "Ready")
		row.Cells = append(row.Cells, s, r, m)
		tableResp.Rows = append(tableResp.Rows, row)
	}

	return tableResp
}

func configurationTableConvert(resp io.Reader) TableResponse {
	var listResp run.ListConfigurationsResponse
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
				Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			},
			{
				Name:        "Service",
				Type:        "string",
				Description: `Custom resource definition column (in JSONPath format): .metadata.labels[serving\.knative\.dev/service]`,
			},
			{
				Name:        "LatestCreated",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.latestCreatedRevisionName",
			},
			{
				Name:        "LatestReady",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.latestReadyRevisionName",
			},
			{
				Name:        "Ready",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].status",
			},
			{
				Name:        "Reason",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].reason",
			},
		},
	}

	for _, item := range listResp.Items {
		row := TableRow{}
		row.Cells = append(row.Cells, item.Metadata.Name)
		row.Cells = append(row.Cells, item.Metadata.Labels["serving.knative.dev/service"])
		row.Cells = append(row.Cells, item.Status.LatestCreatedRevisionName)
		row.Cells = append(row.Cells, item.Status.LatestReadyRevisionName)
		s, r, m := condition(item.Status.Conditions, "Ready")
		row.Cells = append(row.Cells, s, r, m)
		tableResp.Rows = append(tableResp.Rows, row)
	}

	return tableResp
}

func routeTableConvert(resp io.Reader) TableResponse {
	var listResp run.ListRoutesResponse
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
				Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			},
			{
				Name:        "Service",
				Type:        "string",
				Description: `Custom resource definition column (in JSONPath format): .metadata.labels[serving\.knative\.dev/service]`,
			},
			{
				Name:        "URL",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.url",
			},
			{
				Name:        "Ready",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].status",
			},
			{
				Name:        "Reason",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].reason",
			},
		},
	}

	for _, item := range listResp.Items {
		row := TableRow{}
		row.Cells = append(row.Cells, item.Metadata.Name)
		row.Cells = append(row.Cells, item.Metadata.Labels["serving.knative.dev/service"])
		row.Cells = append(row.Cells, item.Status.Url)
		s, r, m := condition(item.Status.Conditions, "Ready")
		row.Cells = append(row.Cells, s, r, m)
		tableResp.Rows = append(tableResp.Rows, row)
	}

	return tableResp
}

func domainMappingTableConvert(resp io.Reader) TableResponse {
	var listResp run.ListDomainMappingsResponse
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
				Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			},
			{
				Name:        "Route",
				Type:        "string",
				Description: `Custom resource definition column (in JSONPath format): .status.mappedRouteName`,
			},
			{
				Name:        "Ready",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].status",
			},
			{
				Name:        "Reason",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].reason",
			},
			{
				Name:        "Message",
				Type:        "string",
				Description: "Custom resource definition column (in JSONPath format): .status.conditions[?(@.type=='Ready')].message",
			},
		},
	}

	for _, item := range listResp.Items {
		row := TableRow{}
		row.Cells = append(row.Cells, item.Metadata.Name)
		row.Cells = append(row.Cells, item.Status.MappedRouteName)
		s, r, m := condition(item.Status.Conditions, "Ready")
		row.Cells = append(row.Cells, s, r, m)
		tableResp.Rows = append(tableResp.Rows, row)
	}

	return tableResp
}

func condition(conds []*run.GoogleCloudRunV1Condition, name string) (status, reason, message interface{}) {
	for _, c := range conds {
		if c.Type == name {
			status = c.Status
			reason = c.Reason
			message = c.Message
			break
		}
	}
	return
}
