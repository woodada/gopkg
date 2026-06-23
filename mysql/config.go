package mysql

type Config struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	DB       string `mapstructure:"db" json:"db" yaml:"db"`
	User     string `mapstructure:"user" json:"user" yaml:"user"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}
