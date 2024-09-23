package vendir

const Filename = "vendir.yml"

type Config struct {
	APIVersion  string `yaml:"apiVersion"`
	Kind        string `yaml:"kind"`
	Directories []struct {
		Path     string             `yaml:"path"`
		Contents []ContentDirectory `yaml:"contents"`
	} `yaml:"directories"`
}

type ContentDirectory struct {
	Git struct {
		URL string `yaml:"url"`
		Ref string `yaml:"ref"`
	} `yaml:"git"`
	Path         string   `yaml:"path"`
	ExcludePaths []string `yaml:"excludePaths"`
}
