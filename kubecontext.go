package main

import (
	"context"

	"google.golang.org/api/run/v1"
	"gopkg.in/yaml.v2"
)

const (
	dummyUser = `dummy`
	ctxPrefix = `cloudrun_`
)

func regions(project string) ([]string, error) {
	ctx := context.TODO()
	svc, err := run.NewService(ctx)
	if err != nil {
		return nil, err
	}
	var out []string
	err = run.NewProjectsService(svc).Locations.List("projects/"+project).Pages(ctx, func(r *run.ListLocationsResponse) error {
		for _, v := range r.Locations {
			out = append(out, v.LocationId)
		}
		return nil
	})
	return out, err
}

type Cluster struct {
	Name    string       `yaml:"name"`
	Cluster ClusterProps `yaml:"cluster"`
}

type ClusterProps struct {
	Server string `yaml:"server"`
}

type User struct {
	Name string `yaml:"name"`
}

type Context struct {
	Context ContextProps `yaml:"context"`
	Name    string       `yaml:"name"`
}

type ContextProps struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace"`
}

type Kubeconfig struct {
	APIVersion     string    `yaml:"apiVersion"`
	Clusters       []Cluster `yaml:"clusters"`
	Contexts       []Context `yaml:"contexts"`
	CurrentContext string    `yaml:"current-context"`
	Kind           string    `yaml:"kind"`
	Users          []User    `yaml:"users"`
}

func mkKubeconfig(project string, regions []string) ([]byte, error) {
	base := `http://localhost:5555`
	kc := Kubeconfig{
		APIVersion: "v1",
		Kind:       "Config",
		Users: []User{
			{Name: dummyUser},
		},
	}
	for _, r := range regions {
		kc.Clusters = append(kc.Clusters, Cluster{
			Name:    r,
			Cluster: ClusterProps{Server: base + "/" + r},
		})
		kc.Contexts = append(kc.Contexts, Context{
			Name: ctxPrefix + r,
			Context: ContextProps{
				Cluster:   r,
				User:      dummyUser,
				Namespace: project,
			},
		})
	}
	kc.CurrentContext = ctxPrefix + "us-central1" // TODO(ahmetb) don't hardcode
	return yaml.Marshal(kc)
}
