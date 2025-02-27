package core

type Module interface {
	Name() string
	CheckRequirements(context map[string]interface{}) (errors []string, warnings []string, successes []string)
}
