package cli

import (
	"strings"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/app/host"
	"github.com/abmpio/configurationx"
	"github.com/abmpio/configurationx/consulv"
	"github.com/spf13/viper"
)

type Host struct {
	app CliApplication
}

type Option func(*Host)

var (
	_host *Host
)

// 安装host环境
func SetupHostEnvironment(companyName string, appName string, version string, opts ...Option) *Host {
	newHost := &Host{}
	host.SetupHostEnvironment(func(hostEnv host.IHostEnvironment) {
		hostEnv.SetAppName(appName)
		hostEnv.SetAppVersion(version)
	})

	c := configurationx.Load(companyName,
		configurationx.ReadFromDefaultPath(),
		configurationx.ReadFromEtcFolder(appName))

	// setup consulPath environment value from default config path
	setupConsulPathHostEnvironment()

	consulPathList := []string{}
	abmpConsulPath := host.GetHostEnvironment().GetEnvString(host.ENV_ConsulPath)
	if len(strings.TrimSpace(abmpConsulPath)) > 0 {
		consulPathList = append(consulPathList, c.Consul.AppendSuffixPathForKVPath(abmpConsulPath))
	} else {
		consulPathList = append(consulPathList, c.Consul.AppendSuffixPathForKVPath("abmpio"))
	}
	envAppNameValue := host.GetHostEnvironment().GetEnvString(host.ENV_AppName)
	if len(envAppNameValue) > 0 {
		consulPathList = append(consulPathList, c.Consul.AppendSuffixPathForKVPath(envAppNameValue))
	}

	_, err := configurationx.Use(consulv.ReadFromConsul(*c.Consul, consulPathList),
		configurationx.ReadFromConfiguration(c))
	if err != nil {
		//panic if configuration error
		panic(err)
	}
	configurationx.GetInstance().UnmarshFromKey("logger", log.DefaultLogConfiguration)

	v := configurationx.GetInstance().GetViper()
	for _, eachKey := range v.AllKeys() {
		isEnvKey := host.IsEnvKey(eachKey)
		if !isEnvKey {
			continue
		}
		value := v.Get(eachKey)
		host.GetHostEnvironment().SetEnv(eachKey, value)
	}
	for _, eachOpt := range opts {
		eachOpt(newHost)
	}
	return newHost
}

func setupConsulPathHostEnvironment() {
	v := viper.New()
	configurationx.SetupViperFromDefaultPath(v)
	keyValue := v.GetString(host.ENV_ConsulPath)
	if len(keyValue) <= 0 {
		return
	}
	// set environment key value
	host.GetHostEnvironment().SetEnv(host.ENV_ConsulPath, keyValue)
}

func (h *Host) Build(cmd ...interface{}) *Host {
	app := newCliApplication(cmd...)
	app.Build()

	app.SystemConfig().App.
		WithName(host.GetHostEnvironment().GetEnvString(host.ENV_AppName)).
		WithVersion(host.GetHostEnvironment().GetEnvString(host.ENV_AppVersion))
	h.app = app

	_host = h
	return h
}

func (h *Host) Run() *Host {
	h.app.Run()
	return h
}

// Get Application
func (h *Host) Application() CliApplication {
	return h.app
}

// get Host instance
func GetHost() *Host {
	return _host
}
