package ui

import (
	"fmt"
	"sort"

	"gopkg.in/AlecAivazis/survey.v1"
	"k8s.io/klog"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/openshift/odo/pkg/util"
)

// SelectStarterProject allows user to select starter project in the prompt
func SelectStarterProject(projects []devfilev1.StarterProject) string {

	if len(projects) == 0 {
		return ""
	}

	projectNames := getProjectNames(projects)

	var download = false
	var selectedProject string
	prompt := &survey.Confirm{Message: "Do you want to download a starter project"}
	err := survey.AskOne(prompt, &download, nil)
	ui.HandleError(err)

	if !download {
		return ""
	}

	// select the only starter project in devfile
	if len(projectNames) == 1 {
		return projectNames[0]
	}

	// If multiple projects present give options to select
	promptSelect := &survey.Select{
		Message: "Which starter project do you want to download",
		Options: projectNames,
	}

	err = survey.AskOne(promptSelect, &selectedProject, survey.Required)
	ui.HandleError(err)
	return selectedProject

}

// SelectDevfileComponentType lets the user to select the devfile component type in the prompt
func SelectDevfileComponentType(options []catalog.DevfileComponentType) string {
	var componentType string
	prompt := &survey.Select{
		Message: "Which devfile component type do you wish to create",
		Options: getDevfileComponentTypeNameCandidates(options),
	}
	err := survey.AskOne(prompt, &componentType, survey.Required)
	ui.HandleError(err)
	return componentType
}

// EnterDevfileComponentName lets the user to specify the component name in the prompt
func EnterDevfileComponentName(defaultComponentName string) string {
	var componentName string
	prompt := &survey.Input{
		Message: "What do you wish to name the new devfile component",
		Default: defaultComponentName,
	}
	err := survey.AskOne(prompt, &componentName, survey.Required)
	ui.HandleError(err)
	return componentName
}

// EnterDevfileComponentProject lets the user to specify the component project in the prompt
func EnterDevfileComponentProject(defaultComponentNamespace string) string {
	var name string
	prompt := &survey.Input{
		Message: "What project do you want the devfile component to be created in",
		Default: defaultComponentNamespace,
	}
	err := survey.AskOne(prompt, &name, validation.NameValidator)
	ui.HandleError(err)
	return name
}

// SelectComponentType lets the user to select the builder image (name only) in the prompt
func SelectComponentType(options []catalog.ComponentType) string {
	var componentType string
	prompt := &survey.Select{
		Message: "Which component type do you wish to create",
		Options: getComponentTypeNameCandidates(options),
	}
	err := survey.AskOne(prompt, &componentType, survey.Required)
	ui.HandleError(err)
	return componentType
}

func getDevfileComponentTypeNameCandidates(options []catalog.DevfileComponentType) []string {
	result := make([]string, len(options))
	for i, option := range options {
		result[i] = option.Name
	}
	sort.Strings(result)
	return result
}

func getProjectNames(projects []devfilev1.StarterProject) []string {
	result := make([]string, len(projects))
	for i, project := range projects {
		result[i] = project.Name
	}
	sort.Strings(result)
	return result
}

func getComponentTypeNameCandidates(options []catalog.ComponentType) []string {
	result := make([]string, len(options))
	for i, option := range options {
		result[i] = option.Name
	}
	sort.Strings(result)
	return result
}

// SelectImageTag lets the user to select a specific tag for the previously selected builder image in a prompt
func SelectImageTag(options []catalog.ComponentType, selectedComponentType string) string {
	var tag string
	prompt := &survey.Select{
		Message: fmt.Sprintf("Which version of '%s' component type do you wish to create", selectedComponentType),
		Options: getTagCandidates(options, selectedComponentType),
	}
	err := survey.AskOne(prompt, &tag, survey.Required)
	ui.HandleError(err)
	return tag
}

func getTagCandidates(options []catalog.ComponentType, selectedComponentType string) []string {
	for _, option := range options {
		if option.Name == selectedComponentType {
			sort.Strings(option.Spec.NonHiddenTags)
			return option.Spec.NonHiddenTags
		}
	}
	klog.V(4).Infof("Selected component type %s was not part of the catalog images", selectedComponentType)
	return []string{}
}

// SelectSourceType lets the user select a specific config.SrcType in a prompty
func SelectSourceType(sourceTypes []config.SrcType) config.SrcType {
	options := make([]string, len(sourceTypes))
	for i, sourceType := range sourceTypes {
		options[i] = fmt.Sprint(sourceType)
	}

	var selectedSourceType string
	prompt := &survey.Select{
		Message: "Which input type do you wish to use for the component",
		Options: options,
	}
	err := survey.AskOne(prompt, &selectedSourceType, survey.Required)
	ui.HandleError(err)

	for _, sourceType := range sourceTypes {
		if selectedSourceType == fmt.Sprint(sourceType) {
			return sourceType
		}
	}
	klog.V(4).Infof("Selected source type %s was not part of the source type options", selectedSourceType)
	return config.NONE
}

// EnterInputTypePath allows the user to specify the path on the filesystem in a prompt
func EnterInputTypePath(inputType string, currentDir string, defaultPath ...string) string {
	var path string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Location of %s component, relative to '%s'", inputType, currentDir),
	}

	if len(defaultPath) == 1 {
		prompt.Default = defaultPath[0]
	}

	err := survey.AskOne(prompt, &path, validation.PathValidator)
	ui.HandleError(err)

	return path
}

// we need this because the validator for the component name needs use info from the Context
// so we effectively return a closure that references the context
func createComponentNameValidator(context *genericclioptions.Context) survey.Validator {
	return func(input interface{}) error {
		if s, ok := input.(string); ok {
			err := validation.ValidateName(s)
			if err != nil {
				return err
			}

			exists, err := component.Exists(context.Client, s, context.Application)
			if err != nil {
				klog.V(4).Info(err)
				return fmt.Errorf("Unable to determine if component '%s' exists or not", s)
			}
			if exists {
				return fmt.Errorf("Component with name '%s' already exists in application '%s'", s, context.Application)
			}

			return nil
		}

		return fmt.Errorf("can only validate strings, got %v", input)
	}
}

// EnterComponentName allows the user to specify the component name in a prompt
func EnterComponentName(defaultName string, context *genericclioptions.Context) string {
	var path string
	prompt := &survey.Input{
		Message: "What do you wish to name the new component",
		Default: defaultName,
	}
	err := survey.AskOne(prompt, &path, createComponentNameValidator(context))
	ui.HandleError(err)
	return path
}

// EnterOpenshiftName allows the user to specify the app name in a prompt
func EnterOpenshiftName(defaultName string, message string, context *genericclioptions.Context) string {
	var name string
	prompt := &survey.Input{
		Message: message,
		Default: defaultName,
	}
	err := survey.AskOne(prompt, &name, validation.NameValidator)
	ui.HandleError(err)
	return name
}

// EnterGitInfo will display two prompts, one of the URL of the project and one of the ref
func EnterGitInfo() (string, string) {
	gitURL := enterGitInputTypePath()
	gitRef := enterGitRef("master")

	return gitURL, gitRef
}

func enterGitInputTypePath() string {
	var path string
	prompt := &survey.Input{
		Message: "What is the URL of the git repository you wish the new component to use",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

func enterGitRef(defaultRef string) string {
	var path string
	prompt := &survey.Input{
		Message: "What git ref (branch, tag, commit) do you wish to use",
		Default: defaultRef,
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

// EnterPorts allows the user to specify the ports to be used in a prompt
func EnterPorts() []string {
	var portsStr string
	prompt := &survey.Input{
		Message: "Enter the ports you wish to set (for example: 8080,8100/tcp,9100/udp). Simply press 'Enter' to avoid setting them",
		Default: "",
	}
	err := survey.AskOne(prompt, &portsStr, validation.PortsValidator)
	ui.HandleError(err)

	return util.GetSplitValuesFromStr(portsStr)
}

// EnterEnvVars allows the user to specify the environment variables to be used in a prompt
func EnterEnvVars() []string {
	var envVarsStr string
	prompt := &survey.Input{
		Message: "Enter the environment variables you would like to set (for example: MY_TYPE=backed,PROFILE=dev). Simply press 'Enter' to avoid setting them",
		Default: "",
	}
	err := survey.AskOne(prompt, &envVarsStr, validation.KeyEqValFormatValidator)
	ui.HandleError(err)

	return util.GetSplitValuesFromStr(envVarsStr)
}

// EnterMemory allows the user to specify the memory limits to be used in a prompt
func EnterMemory(typeStr string, defaultValue string) string {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter the %s memory (for example 100Mi)", typeStr),
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result, survey.Required)
	ui.HandleError(err)

	return result
}

// EnterCPU allows the user to specify the cpu limits to be used in a prompt
func EnterCPU(typeStr string, defaultValue string) string {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter the %s CPU (for example 100m or 2)", typeStr),
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result, survey.Required)
	ui.HandleError(err)

	return result
}
