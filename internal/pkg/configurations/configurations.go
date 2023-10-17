package configurations

import (
	"github.com/kelseyhightower/envconfig"
	gerrors "clusterviz/internal/pkg/gerror"
)


type ServiceConfigurations struct {
	LogLevel           string `envconfig:"LOG_LEVEL" default:"info"`
	Port               string `envconfig:"PORT" default:"9090"`
	HeaderReadTimeout int
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING"`
}

func GetServiceConfigurations() (serviceConf *ServiceConfigurations, err error) {
	serviceConf = &ServiceConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, gerrors.NewFromError(gerrors.ServiceSetup, err)
	}
	return
}
