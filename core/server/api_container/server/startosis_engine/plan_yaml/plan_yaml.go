package plan_yaml

import (
	"github.com/go-yaml/yaml"
)

const (
	HTTP ApplicationProtocol = "HTTP"
	UDP  TransportProtocol   = "UDP"
	TCP  TransportProtocol   = "TCP"

	SHELL  TaskType = "sh"
	PYTHON TaskType = "python"
	EXEC   TaskType = "exec"
)

// TODO: there's really no point in making any of these references, consider just making them copies
type PlanYaml struct {
	PackageId      string           `yaml:"packageId,omitempty"`
	Services       []*Service       `yaml:"services,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"`
	Tasks          []*Task          `yaml:"tasks,omitempty"`
}

// Service represents a service in the system.
type Service struct {
	Uuid       string                 `yaml:"uuid,omitempty"`       // done
	Name       string                 `yaml:"name,omitempty"`       // done
	Image      *ImageSpec             `yaml:"image,omitempty"`      // done
	Cmd        []string               `yaml:"command,omitempty"`    // done
	Entrypoint []string               `yaml:"entrypoint,omitempty"` // done
	EnvVars    []*EnvironmentVariable `yaml:"envVars,omitempty"`    // done
	Ports      []*Port                `yaml:"ports,omitempty"`      // done
	Files      []*FileMount           `yaml:"files,omitempty"`      // done

	// TODO: support remaining fields in the ServiceConfig
}

type ImageSpec struct {
	ImageName string `yaml:"name,omitempty"`

	// for built images
	BuildContextLocator string `yaml:"buildContextLocator,omitempty"`
	TargetStage         string `yaml:"targetStage,omitempty"`

	// for images from registry
	Registry string `yaml:"registry,omitempty"`
}

// FilesArtifact represents a collection of files.
type FilesArtifact struct {
	Uuid  string   `yaml:"uuid,omitempty"`
	Name  string   `yaml:"name,omitempty"`
	Files []string `yaml:"files,omitempty"`
}

// EnvironmentVariable represents an environment variable.
type EnvironmentVariable struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// Port represents a port.
type Port struct {
	Name   string `yaml:"name,omitempty"`
	Number uint16 `yaml:"number,omitempty"`

	TransportProtocol   TransportProtocol   `yaml:"transportProtocol,omitempty"`
	ApplicationProtocol ApplicationProtocol `yaml:"applicationProtocol,omitempty"`
}

// ApplicationProtocol represents the application protocol used.
type ApplicationProtocol string

// TransportProtocol represents transport protocol used.
type TransportProtocol string

// FileMount represents a mount point for files.
type FileMount struct {
	MountPath      string           `yaml:"mountPath,omitempty"`
	FilesArtifacts []*FilesArtifact `yaml:"filesArtifacts,omitempty"` // TODO: support persistent directories
}

// Task represents a task to be executed.
type Task struct {
	Uuid     string           `yaml:"uuid,omitempty"`     // done
	Name     string           `yaml:"name,omitempty"`     // done
	TaskType TaskType         `yaml:"taskType,omitempty"` // done
	RunCmd   []string         `yaml:"command,omitempty"`  // done
	Image    string           `yaml:"image,omitempty"`    // done
	Files    []*FileMount     `yaml:"files,omitempty"`    // done
	Store    []*FilesArtifact `yaml:"store,omitempty"`    // done

	// only exists on SHELL tasks
	EnvVars []*EnvironmentVariable `yaml:"envVar,omitempty"` // done

	// only exists on PYTHON tasks
	PythonPackages []string `yaml:"pythonPackages,omitempty"`
	PythonArgs     []string `yaml:"pythonArgs,omitempty"`

	// service name
	ServiceName     string  `yaml:"serviceName,omitempty"`
	AcceptableCodes []int64 `yaml:"acceptableCodes,omitempty"`
}

// TaskType represents the type of task (either PYTHON or SHELL)
type TaskType string

func CreateEmptyPlan(packageId string) *PlanYaml {
	return &PlanYaml{
		PackageId:      packageId,
		Services:       nil,
		FilesArtifacts: nil,
		Tasks:          nil,
	}
}

func (pyg PlanYaml) GenerateYaml() (string, error) {
	yamlBytes, err := yaml.Marshal(pyg)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}
