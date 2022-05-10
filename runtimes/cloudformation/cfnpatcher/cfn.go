package cfnpatcher

import (
	"context"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	Kilt               string
	ImageAuthSecret    string
	OptIn              bool
	RecipeConfig       string
	UseRepositoryHints bool
	LogGroup           string
}

type InstrumentationHints struct {
	IgnoreContainersNamed  []string
	IncludeContainersNamed []string
	HasGlobalInclude       bool
}

var Empty struct{}

const KiltIgnoreTag = "kilt-ignore"
const KiltIncludeTag = "kilt-include"
const KiltIgnoreContainersTag = "kilt-ignore-containers"
const KiltIncludeContainersTag = "kilt-include-containers"

func isIgnored(tags map[string]string, isOptIn bool) bool {
	_, included := tags[KiltIncludeTag]
	_, ignored := tags[KiltIgnoreTag]
	_, hasNamedContainerIncluded := tags[KiltIncludeContainersTag]

	return !((isOptIn && (included || hasNamedContainerIncluded)) || (!isOptIn && !ignored))
}

func extractContainersFromTag(tags map[string]string, tag string) []string {
	containers := make([]string, 0)
	containerList, hasIgnores := tags[tag]
	if hasIgnores {
		containers = strings.Split(containerList, ",")
	}
	return containers
}

func extractHintsFromTags(tags map[string]string) *InstrumentationHints {
	_, included := tags[KiltIncludeTag]
	return &InstrumentationHints{
		IgnoreContainersNamed:  extractContainersFromTag(tags, KiltIgnoreContainersTag),
		IncludeContainersNamed: extractContainersFromTag(tags, KiltIncludeContainersTag),
		HasGlobalInclude:       included,
	}
}

func Patch(ctx context.Context, configuration *Configuration, fragment []byte) ([]byte, error) {
	l := log.Ctx(ctx)
	template, err := gabs.ParseJSON(fragment)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse input fragment")
		return nil, err
	}

	taskDefinitionToResource := make(map[string]*gabs.Container)
	iamRoleToResource := make(map[string]*gabs.Container)

	for name, resource := range template.S("Resources").ChildrenMap() {
		if matchFargate(resource) {
			taskDefinitionToResource[name] = resource
		}
		if isIamRole(resource) {
			iamRoleToResource[name] = resource
		}
	}

	for name, resource := range taskDefinitionToResource {
		tags := getTags(resource)

		if isIgnored(tags, configuration.OptIn) {
			l.Info().Str("resource", name).Msg("ignored resource due to tag")
			continue
		}
		l.Info().Str("resource", name).Msg("patching task definition")
		if err != nil {
			l.Error().Err(err).Str("resource", name).Msg("could not generate kilt instructions")
			continue
		}

		executionRole, ok := getExecutionRole(resource)
		if !ok {
			l.Warn().Str("resource", name).Msg("no execution role was specified")
		}

		hints := extractHintsFromTags(tags)
		_, err = applyTaskDefinitionPatch(ctx, name, resource, configuration, hints, iamRoleToResource, executionRole)
		if err != nil {
			l.Error().Err(err).Str("resource", name).Msgf("could not patch resource")
		}
	}

	return template.Bytes(), nil
}
