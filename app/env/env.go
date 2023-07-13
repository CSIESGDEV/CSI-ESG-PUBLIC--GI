package env

import (
	"reflect"

	"github.com/caarlos0/env/v6"
)

// Config :
var Config = struct {
	App struct {
		SystemPath      string `env:"SYSTEM_HOST_PATH,required"`
		AdminPortalPath string `env:"ADMIN_PORTAL_PATH,required"`
		UserPortalPath  string `env:"USER_PORTAL_PATH,required"`
		Env             string `env:"ENV,required"`
	}
	Storage struct {
		Path string `env:"STORAGE_PATH,required"`
	}
	Jwt struct {
		Secret string `env:"JWT_SECRET,required"`
		Issuer string `env:"JWT_ISSUER,required"`
	}
	Mongo struct {
		Host       string `env:"MONGO_DB_HOST,required"`
		TestDBName string `env:"MONGO_TEST_DB_NAME,required"`
		DBName     string `env:"MONGO_DB_NAME,required"`
		Username   string `env:"MONGO_DB_USER,required"`
		Password   string `env:"MONGO_DB_PASS,required"`
	}
	AWS struct {
		Sender  string `env:"AWS_SES_SENDER,required"`
		Sender2 string `env:"AWS_SES_SENDER_2,required"`
		Profile string `env:"AWS_PROFILE,required"`
	}
	S3 struct {
		Region    string `env:"AWS_REGION"`
		KeyID     string `env:"AWS_ACCESS_KEY_ID"`
		AccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
		S3BUCKET  string `env:"AWS_S3_BUCKET"`
	}
}{}

func init() {
	if err := env.ParseWithFuncs(&Config,
		map[reflect.Type]env.ParserFunc{});
		err != nil {
		panic(err)
	}
}

// IsProduction : Return true if the environment is production
var IsProduction = func() bool {
	return Config.App.Env == "production"
}

// IsSandbox : Return true if the environment is sandbox
var IsSandbox = func() bool {
	return Config.App.Env == "sandbox"
}

// IsDevelopment : Return true if the environment is development
var IsDevelopment = func() bool {
	return Config.App.Env == "development"
}
