package config

type Gitlab struct {
	GitlabToken  string `env:",required,notEmpty"`
	GitlabAPIURL string `env:"GITLAB_API_URL,required,notEmpty"`
}

type Config struct {
	Port            string `env:"PORT" envDefault:"3000"`
	ApplicationName string `env:"APP_NAME" envDefault:"Assignment Service"`
	Gitlab          Gitlab
	HiveURL         string `env:"HIVE_URL,required,notEmpty"`
	InsecureHive    bool   `env:"INSECURE_HIVE" envDefault:"false"`
	TemplatesPath   string `env:"TEMPLATES_PATH,required,notEmpty"`
}
