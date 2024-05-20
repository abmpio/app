package cli

import (
	"strings"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/app"
	"github.com/abmpio/app/host"
	"github.com/abmpio/configurationx"
	"github.com/abmpio/configurationx/consulv"
)

type Host struct {
	app app.Application
}

type Option func(*Host)

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

	consulPathList := []string{}
	abmpConsulPath := host.GetHostEnvironment().GetEnvString(host.ENV_ConsulPath)
	if len(strings.TrimSpace(abmpConsulPath)) > 0 {
		consulPathList = append(consulPathList, abmpConsulPath)
	} else {
		consulPathList = append(consulPathList, "abmp")
	}
	envAppNameValue := host.GetHostEnvironment().GetEnvString(host.ENV_AppName)
	if len(envAppNameValue) > 0 {
		consulPathList = append(consulPathList, envAppNameValue)
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

func (h *Host) Build(cmd ...interface{}) *Host {
	app := NewCliApplication(cmd...).Build()
	app.SystemConfig().App.
		WithName(host.GetHostEnvironment().GetEnvString(host.ENV_AppName)).
		WithVersion(host.GetHostEnvironment().GetEnvString(host.ENV_AppVersion))
	h.app = app
	return h
}

func (h *Host) Run() *Host {
	h.app.Run()
	return h
}
