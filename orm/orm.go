package orm

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/QunQunLab/ego/conf"
	"github.com/QunQunLab/ego/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var (
	mutex   sync.RWMutex
	engines = map[string]*xorm.EngineGroup{}
)

type masterOption struct {
	Host         string `json:"mysql_master:host"`
	User         string `json:"mysql_master:user"`
	Password     string `json:"mysql_master:password"`
	Database     string `json:"mysql_master:database"`
	Charset      string `json:"mysql_master:charset"`
	MaxOpenConns int    `json:"mysql_master:max_open_conns"`
	MaxIdleConns int    `json:"mysql_master:max_idle_conns"`
	Debug        bool   `json:"mysql_master:debug"`
}

type slaveOption struct {
	Host         string `json:"mysql_slave:host"`
	User         string `json:"mysql_slave:user"`
	Password     string `json:"mysql_slave:password"`
	Database     string `json:"mysql_slave:database"`
	Charset      string `json:"mysql_slave:charset"`
	MaxOpenConns int    `json:"mysql_slave:max_open_conns"`
	MaxIdleConns int    `json:"mysql_slave:max_idle_conns"`
	Debug        bool   `json:"mysql_slave:debug"`
}

func getEngine(mysqlConf string) *xorm.Engine {
	if mysqlConf == "mysql_master" {
		master := masterOption{}
		err := conf.Unmarshal(&master)
		if err != nil {
			log.Error("%v unmarshal err:%v", mysqlConf, err)
		}
		// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
		mds := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=%v", master.User, master.Password, master.Host, master.Database, master.Charset)
		log.Trace("%v source:%v", mysqlConf, mds)
		me, err := xorm.NewEngine("mysql", mds)
		if err != nil {
			log.Error("new %v engine err:%v", mysqlConf, err)
		}
		if master.MaxIdleConns > 0 {
			me.SetMaxIdleConns(master.MaxIdleConns)
		}
		if master.MaxOpenConns > 0 {
			me.SetMaxOpenConns(master.MaxOpenConns)
		}
		me.ShowSQL(master.Debug)
		return me
	}

	if mysqlConf == "mysql_slave" {
		slave := slaveOption{}
		err := conf.Unmarshal(&slave)
		if err != nil {
			panic(err)
		}
		sds := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=%v", slave.User, slave.Password, slave.Host, slave.Database, slave.Charset)
		log.Trace("%v source:%v", mysqlConf, sds)
		se, err := xorm.NewEngine("mysql", sds)
		if err != nil {
			log.Fatal("new %v engine err:%v", mysqlConf, err)
		}
		if slave.MaxIdleConns > 0 {
			se.SetMaxIdleConns(slave.MaxIdleConns)
		}
		if slave.MaxOpenConns > 0 {
			se.SetMaxOpenConns(slave.MaxOpenConns)
		}
		se.ShowSQL(slave.Debug)
		return se
	}

	section := conf.Get(mysqlConf)
	host, _ := section.String("host")
	user, _ := section.String("user")
	pass, _ := section.String("password")
	database, _ := section.String("database")
	charset, _ := section.String("charset")
	debug, _ := section.Bool("debug")
	idleConn, _ := section.Int("max_idle_conns")
	openConn, _ := section.Int("max_open_conns")
	sds := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=%v", user, pass, host, database, charset)
	log.Trace("slave data source:%v", sds)
	se, err := xorm.NewEngine("mysql", sds)
	if err != nil {
		log.Fatal("new engine:%v err:%v", mysqlConf, err)
	}
	if idleConn > 0 {
		se.SetMaxIdleConns(int(idleConn))
	}
	if openConn > 0 {
		se.SetMaxOpenConns(int(openConn))
	}
	se.ShowSQL(debug)
	return se
}

func Engine(c ...string) *xorm.EngineGroup {
	m := "mysql_master"
	s := "mysql_slave"

	if len(c) >= 1 {
		m = c[0]
		s = c[1]
	}

	// gen key with first two param
	buf := bytes.NewBufferString(m)
	buf.WriteString(s)
	h := md5.New()
	h.Write(buf.Bytes())
	key := hex.EncodeToString(h.Sum(nil))

	if val, ok := engines[key]; ok {
		log.Trace("get orm engine:%v", key)
		return val
	} else {
		log.Trace("new orm engine:%v", key)
		master := getEngine(m)

		// slaves
		var slaves []*xorm.Engine
		if len(c) <= 0 {
			slaves = append(slaves, getEngine(s))
		} else {
			for i := 1; i < len(c); i++ {
				slaves = append(slaves, getEngine(c[i]))
			}
		}

		eg, err := xorm.NewEngineGroup(master, slaves)
		if err != nil {
			log.Fatal("engineGroup:%v err:%v", c, err)
			panic(err)
		}

		mutex.Lock()
		engines[key] = eg
		mutex.Unlock()
		return eg
	}
}
