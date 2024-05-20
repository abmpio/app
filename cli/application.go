package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/abmpio/abmp/pkg/log"
	"github.com/abmpio/app"
)

// 配置服务
type ServiceConfigurator func(CliApplication)

var (
	_registedConfiguratorList []ServiceConfigurator
	_syncOnce                 sync.Once
	Application               CliApplication
)

// web应用
type CliApplication interface {
	app.Application

	GetServiceProvider() app.IServiceProvider
	ConfigureService()
}

type defaultCliApplication struct {
	app.BaseApplication
	root Command

	serviceProvider app.IServiceProvider
}

type CommandNameValue struct {
	Name    string
	Command interface{}
}

const (
	// RootCommandName the instance name of cli.rootCommand
	RootCommandName = "cli.rootCommand"
)

// new一个cli应用
func NewCliApplication(cmd ...interface{}) CliApplication {
	newApp := &defaultCliApplication{}
	newApp.initialize(cmd...)
	if app.HostApplication != nil {
		newApp.serviceProvider = app.HostApplication.GetServiceProvider()
	}
	return newApp
}

// 设置Application属性值
func SetCliApplication(cliApp CliApplication) {
	Application = cliApp
}

func (a *defaultCliApplication) GetServiceProvider() app.IServiceProvider {
	return a.serviceProvider
}

func (a *defaultCliApplication) initialize(cmd ...interface{}) (err error) {
	if len(cmd) > 0 {
		app.Register(RootCommandName, cmd[0])
	}
	err = a.Initialize()
	return
}

// 构建应用运行所需的环境
func (a *defaultCliApplication) Build() app.Application {
	//先调用基类的构建函数
	a.BaseApplication.Build()

	basename := filepath.Base(os.Args[0])
	basename = strings.ToLower(basename)
	basename = strings.TrimSuffix(basename, ".exe")

	f := a.ConfigurableFactory()
	f.SetInstance(app.ApplicationContextName, a)

	// 处理自动注入配置
	a.BuildConfigurations()

	// cli root command
	r := f.GetInstance(RootCommandName)
	var root Command
	if r != nil {
		root = r.(Command)
		Register(root)
		a.root = root
		root.EmbeddedCommand().Use = basename
	}

	a.AfterInitialization()
	return a
}

func (a *defaultCliApplication) ConfigureService() {
	_syncOnce.Do(func() {
		for _, eachOption := range _registedConfiguratorList {
			configuratorName := getServiceConfiguratorTypeName(eachOption)
			if !app.HostApplication.SystemConfig().App.IsRunInCli {
				log.Logger.Info(fmt.Sprintf("begin run ServiceConfigurator,%s", configuratorName))
			}
			eachOption(a)
			if !app.HostApplication.SystemConfig().App.IsRunInCli {
				log.Logger.Info(fmt.Sprintf("finish run ServiceConfigurator,%s", configuratorName))
			}
		}
	})
}

func getServiceConfiguratorTypeName(configuratorFunc ServiceConfigurator) string {
	if configuratorFunc == nil {
		return ""
	}
	return runtime.FuncForPC(reflect.ValueOf(configuratorFunc).Pointer()).Name()
}

// 配置服务
func ConfigureService(opts ...ServiceConfigurator) {
	_registedConfiguratorList = append(_registedConfiguratorList, opts...)
}

// 设置应用属性名
func (a *defaultCliApplication) SetProperty(name string, value ...interface{}) app.Application {
	a.BaseApplication.SetProperty(name, value...)
	return a
}

func (a *defaultCliApplication) SetAddCommandLineProperties(enabled bool) app.Application {
	a.BaseApplication.SetAddCommandLineProperties(enabled)
	return a
}

// 初始化应用
func (a *defaultCliApplication) Initialize() error {
	return a.BaseApplication.Initialize()
}

// 运行应用
func (a *defaultCliApplication) Run() {
	if a.root != nil {
		a.root.Exec()
	}
	a.Shutdown()
}
