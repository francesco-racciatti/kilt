package cfnpatcher

import "github.com/Jeffail/gabs/v2"

func matchFargate(resource *gabs.Container) bool {
	return isTaskDefinition(resource) && requiresFargate(resource)
}

func requiresFargate(resource *gabs.Container) bool {
	exists := resource.Exists("Properties", "RequiresCompatibilities")
	if !exists {
		return false
	}
	for _, value := range resource.S("Properties", "RequiresCompatibilities").Children() {
		compat, ok := value.Data().(string)
		if ok && compat == "FARGATE" {
			return true
		}
	}
	return false
}

func isTaskDefinition(resource *gabs.Container) bool {
	return hasType(resource, "AWS::ECS::TaskDefinition")
}

func isIamRole(resource *gabs.Container) bool {
	return hasType(resource, "AWS::IAM::Role")
}

func getExecutionRole(resource *gabs.Container) (string, bool) {
	k, ok := resource.S("Properties", "ExecutionRoleArn", "Ref").Data().(string)
	return k, ok
}

func hasType(resource *gabs.Container, resourceType string) bool {
	exists := resource.Exists("Type")
	if !exists {
		return false
	}
	value, ok := resource.S("Type").Data().(string)
	return ok && value == resourceType
}
