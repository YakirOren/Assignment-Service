package config

type Gitlab struct {
	GitlabToken  string `env:",required,notEmpty"`
	GitlabAPIURL string `env:"GITLAB_API_URL,required,notEmpty"`
}

type Config struct {
	Port            string `env:"PORT" envDefault:"3000"`
	ApplicationName string `env:"APP_NAME" envDefault:"On New Assignment Service"`
	Gitlab          Gitlab
}
