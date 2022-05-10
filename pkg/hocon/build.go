package hocon

import (
	"fmt"

	"github.com/go-akka/configuration"

	"github.com/falcosecurity/kilt/pkg/kilt"
)

func extractBuild(config *configuration.Config) (*kilt.Build, error) {
	b := new(kilt.Build)

	b.Image = config.GetString("build.image")
	b.EntryPoint = config.GetStringList("build.entry_point")
	if b.EntryPoint == nil {
		b.EntryPoint = make([]string, 0)
	}
	b.Command = config.GetStringList("build.command")
	if b.Command == nil {
		b.Command = make([]string, 0)
	}

	b.EnvironmentVariables = extractToStringMap(config, "build.environment_variables")

	if config.IsArray("build.mount") {
		mounts := config.GetValue("build.mount").GetArray()

		for k, m := range mounts {
			if m.IsObject() {
				mount := m.GetObject()

				resource := kilt.BuildResource{
					Name:       mount.GetKey("name").GetString(),
					Image:      mount.GetKey("image").GetString(),
					Volumes:    mount.GetKey("volumes").GetStringList(),
					EntryPoint: mount.GetKey("entry_point").GetStringList(),
				}

				if resource.Image == "" || len(resource.Volumes) == 0 || len(resource.EntryPoint) == 0 {
					return nil, fmt.Errorf("error at build.mount.%d: image, volumes and entry_point are all required ", k)
				}

				b.Resources = append(b.Resources, resource)
			}
		}
	}

	if config.IsArray("build.execution_policies") {
		policies := config.GetValue("build.execution_policies").GetArray()

		for k, m := range policies {
			if m.IsObject() {
				policy := m.GetObject()

				resource := kilt.PolicyResource{
					Name:     policy.GetKey("name").GetString(),
					Version:  policy.GetKey("version").GetString(),
					Effect:   policy.GetKey("effect").GetString(),
					Action:   policy.GetKey("action").GetStringList(),
					Resource: policy.GetKey("resource").GetStringList(),
				}

				if resource.Version == "" || resource.Effect == "" || len(resource.Action) == 0 || len(resource.Resource) == 0 {
					return nil, fmt.Errorf("error at build.execution_policies.%d: version, effect, action, and resource are all required ", k)
				}

				b.ExecutionPolicies = append(b.ExecutionPolicies, resource)
			}
		}
	}

	return b, nil
}
