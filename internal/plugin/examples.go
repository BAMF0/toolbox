package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bamf0/toolbox/internal/config"
)

// DockerPlugin is an example plugin that adds Docker context support
type DockerPlugin struct {
	name    string
	version string
}

// NewDockerPlugin creates a new Docker plugin instance
func NewDockerPlugin() *DockerPlugin {
	return &DockerPlugin{
		name:    "docker",
		version: "1.0.0",
	}
}

// Name returns the plugin identifier
func (p *DockerPlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *DockerPlugin) Version() string {
	return p.version
}

// Contexts returns the Docker-related contexts
func (p *DockerPlugin) Contexts() map[string]config.ContextConfig {
	return map[string]config.ContextConfig{
		"docker": {
			Commands: map[string]string{
				"build":   "docker build -t $(basename $(pwd)) .",
				"run":     "docker run -it $(basename $(pwd))",
				"push":    "docker push $(basename $(pwd))",
				"compose": "docker-compose up",
				"stop":    "docker-compose down",
				"logs":    "docker-compose logs -f",
				"shell":   "docker exec -it $(docker ps -q -f name=$(basename $(pwd))) /bin/bash",
			},
		},
		"docker-compose": {
			Commands: map[string]string{
				"up":    "docker-compose up -d",
				"down":  "docker-compose down",
				"logs":  "docker-compose logs -f",
				"build": "docker-compose build",
				"restart": "docker-compose restart",
			},
		},
	}
}

// Detect checks if the current directory is a Docker project
func (p *DockerPlugin) Detect(dir string) (string, bool) {
	// Check for Dockerfile
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if fileExists(dockerfilePath) {
		return "docker", true
	}

	// Check for docker-compose.yml or docker-compose.yaml
	composeYml := filepath.Join(dir, "docker-compose.yml")
	composeYaml := filepath.Join(dir, "docker-compose.yaml")
	
	if fileExists(composeYml) || fileExists(composeYaml) {
		return "docker-compose", true
	}

	return "", false
}

// Validate performs plugin validation
func (p *DockerPlugin) Validate() error {
	if p.name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	
	if p.version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	// Validate all commands are non-empty
	for ctxName, ctxConfig := range p.Contexts() {
		if len(ctxConfig.Commands) == 0 {
			return fmt.Errorf("context %q has no commands", ctxName)
		}
		
		for cmdName, cmd := range ctxConfig.Commands {
			if cmd == "" {
				return fmt.Errorf("context %q, command %q is empty", ctxName, cmdName)
			}
		}
	}

	return nil
}

// KubernetesPlugin is an example plugin for Kubernetes
type KubernetesPlugin struct {
	name    string
	version string
}

// NewKubernetesPlugin creates a new Kubernetes plugin
func NewKubernetesPlugin() *KubernetesPlugin {
	return &KubernetesPlugin{
		name:    "kubernetes",
		version: "1.0.0",
	}
}

func (p *KubernetesPlugin) Name() string {
	return p.name
}

func (p *KubernetesPlugin) Version() string {
	return p.version
}

func (p *KubernetesPlugin) Contexts() map[string]config.ContextConfig {
	return map[string]config.ContextConfig{
		"kubernetes": {
			Commands: map[string]string{
				"apply":   "kubectl apply -f .",
				"delete":  "kubectl delete -f .",
				"get":     "kubectl get all",
				"logs":    "kubectl logs -f",
				"describe": "kubectl describe",
				"exec":    "kubectl exec -it",
				"port-forward": "kubectl port-forward",
			},
		},
		"helm": {
			Commands: map[string]string{
				"install":  "helm install",
				"upgrade":  "helm upgrade",
				"rollback": "helm rollback",
				"list":     "helm list",
				"delete":   "helm delete",
			},
		},
	}
}

func (p *KubernetesPlugin) Detect(dir string) (string, bool) {
	// Check for Kubernetes manifests
	manifestFiles := []string{
		"deployment.yaml",
		"deployment.yml",
		"k8s/deployment.yaml",
		"kubernetes/deployment.yaml",
	}

	for _, manifest := range manifestFiles {
		if fileExists(filepath.Join(dir, manifest)) {
			return "kubernetes", true
		}
	}

	// Check for Helm chart
	if fileExists(filepath.Join(dir, "Chart.yaml")) {
		return "helm", true
	}

	return "", false
}

func (p *KubernetesPlugin) Validate() error {
	if p.name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	
	if p.version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	for ctxName, ctxConfig := range p.Contexts() {
		if len(ctxConfig.Commands) == 0 {
			return fmt.Errorf("context %q has no commands", ctxName)
		}
	}

	return nil
}

// Helper function
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
