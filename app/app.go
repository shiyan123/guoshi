package app

import (
	"fmt"
	"guoshi/common/db"
	"path"
	"runtime"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type Application struct {
	ProcParams *ProcParams
	Config     *Config
	GcMysql    *xorm.Engine
	GcRoMysql  *xorm.Engine
	Mongo      *db.MongoDB
}

var (
	appOnce sync.Once
	app     *Application
)

func GetApp() *Application {
	appOnce.Do(func() {
		app = new(Application)
	})

	return app
}

func (a *Application) Prepare() error {
	runtime.GOMAXPROCS(runtime.NumCPU()) //todo

	if err := a.initParams(); err != nil {
		return err
	}

	if err := a.initConfig(a.ProcParams.ConfigPath, a.ProcParams.Environment); err != nil {
		return err
	}

	if err := a.initDb(); err != nil {
		return err
	}

	return nil
}

func (a *Application) initParams() error {
	params, err := LoadParams()
	if err != nil {
		return err
	}

	a.ProcParams = params
	return nil
}

func (a *Application) initConfig(configPath, env string) error {
	config, err := LoadConfig(path.Join(configPath, fmt.Sprintf("config_%s.json", env)))
	if err != nil {
		return err
	}

	a.Config = config
	return nil
}

func (a *Application) initDb() error {

	Mongo, err := db.InitMongo(a.Config.MongoUrl, "guoshi", 30)
	if err != nil {
		err = fmt.Errorf("mongo connection err: %v", err)
		return err
	}

	a.Mongo = Mongo
	return nil
}
