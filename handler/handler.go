package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liuyp5181/base/cache"
	"github.com/liuyp5181/base/etcd"
	"github.com/liuyp5181/base/log"
	"github.com/liuyp5181/base/service"
	"github.com/liuyp5181/base/service/extend"
	"github.com/liuyp5181/base/util"
	"github.com/liuyp5181/configmgr/client"
	"github.com/liuyp5181/gateway/data"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const SessionKeyPrefix = "session"
const UserPowerKeyPrefix = "user_power"

type Config struct {
	Limit  rate.Limit `mapstructure:"limit"`
	Bucket int        `mapstructure:"bucket"`
}

func login(c *gin.Context) {
	fmt.Println("proxy login in")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("ReadAll err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "body is err"})
		return
	}

	etcd.PrintService()

	cl, err := service.GetClient("login.Greeter")
	if err != nil {
		log.Error("GetClient err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "get client is err"})
		return
	}

	trid := util.GenerateId("trace_id", body)
	e := extend.New()
	e.SetClient("trace_id", trid)
	log.Info(c.Request.RemoteAddr, trid, c.Request.URL.Path)

	rsp, err := cl.Proxy(e.Ctx, "Login", body, e.LoadServer())
	if err != nil {
		log.Error("Proxy err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "proxy is err"})
		return
	}

	uid := e.GetServer("user_id")
	session := util.GenerateId("sess", uid)

	key := fmt.Sprintf("%s:%s", SessionKeyPrefix, session)
	r := cache.GetRedis("test")
	_, err = r.Set(e.Ctx, key, uid, time.Second*1200).Result()
	if err != nil {
		log.Error("Set err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "call is err"})
		return
	}

	fmt.Println("proxy login out")

	c.SetCookie("session", session, 1200, "/", "*", false, true)

	c.Data(http.StatusOK, "application/json", rsp)
}

func verify(c *gin.Context) bool {
	serviceName := c.GetString("service_name")
	method := c.GetString("method")

	log.Info(serviceName, method)
	external.Range(func(key, value interface{}) bool {
		log.Info(key, value)
		return true
	})

	e, ok := external.Load(serviceName)
	if !ok {
		log.Error("path err path =", c.Request.URL.Path)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "path is err"})
		return false
	}
	m := e.(map[string]*data.External)
	ext, ok := m[method]
	if !ok {
		log.Error("path err path =", c.Request.URL.Path)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "path is err"})
		return false
	}
	pathPower := ext.Power

	cookie, err := c.Request.Cookie("session")
	if err != nil {
		log.Error("Cookie err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "get cookie fail"})
		return false
	}

	r := cache.GetRedis("test")
	key := fmt.Sprintf("%s:%s", SessionKeyPrefix, cookie.Value)
	uid, err := r.Get(c, key).Result()
	if err != nil {
		log.Error("Get err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "get cache fail"})
		return false
	}

	upKey := fmt.Sprintf("%s:%s", UserPowerKeyPrefix, uid)
	vals, err := r.HMGet(c, upKey, "power", serviceName+"/"+method, serviceName+"/*").Result()
	log.Info(vals, err, vals == nil)
	if err != nil {
		log.Error("HMGet err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "get cache fail"})
		return false
	}

	var redisNil int
	var power int32
	var path, pathAll string
	for i, v := range vals {
		if v == nil {
			redisNil++
			continue
		}
		switch i {
		case 0:
			ai, _ := strconv.Atoi(v.(string))
			power = int32(ai)
		case 1:
			path = v.(string)
		case 2:
			pathAll = v.(string)
		}
	}

	if redisNil == 3 {
		up, err := data.QueryUserPower(uid)
		if err != nil {
			log.Error("QueryUserPower err =", err)
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "call is err"})
			return false
		}
		var vm = make(map[string]string)
		var vs = []interface{}{"power", up.Power}
		for _, p := range strings.Split(up.Path, ",") {
			if len(p) == 0 {
				continue
			}
			vm[p] = "1"
			vs = append(vs, p, "1")
		}
		_, err = r.HMSet(c, upKey, vs...).Result()
		if err != nil {
			log.Error("HMSet err =", err)
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "call is err"})
			return false
		}
		power, path, pathAll = up.Power, vm[serviceName+"/"+method], vm[serviceName+"/*"]
	}

	if len(uid) == 0 {
		log.Error("session is fail, session =", cookie.Value)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "session is fail"})
		return false
	}
	c.Set("user_id", uid)

	if power > pathPower || len(path) > 0 || len(pathAll) > 0 {
		return true
	}

	log.Error("permission too low", power, pathPower, path, pathAll)
	c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "permission too low"})
	return false
}

func transmit(c *gin.Context) {
	fmt.Println("transmit")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("ReadAll err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "body is err"})
		return
	}

	serviceName := c.GetString("service_name") + ".Greeter"
	method := c.GetString("method")

	cl, err := service.GetClient(serviceName)
	if err != nil {
		log.Errorf("GetClient err = %v", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "get client is err"})
		return
	}

	trid := util.GenerateId("trace_id", body)
	e := extend.New()
	e.SetClient("trace_id", trid)
	e.SetClient("user_id", c.GetString("user_id"))
	log.Info(c.Request.RemoteAddr, trid, c.Request.URL.Path)

	//md := metadata.Pairs("trace_id", "123")
	//ctx := metadata.NewOutgoingContext(extend.Background(), md)
	//ctx = metadata.AppendToOutgoingContext(ctx, "user_id", c.GetString("user_id"))
	rsp, err := cl.Proxy(e.Ctx, method, body)
	if err != nil {
		log.Error("Proxy err =", err)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "proxy is err"})
		return
	}

	c.Data(http.StatusOK, "application/json", rsp)
}

func Proxy(c *gin.Context) {
	cfg := client.GetConfig("gateway.yaml").(*Config)
	log.Info("rate", cfg.Limit, cfg.Bucket)
	limiter := rate.NewLimiter(cfg.Limit, cfg.Bucket)
	if !limiter.Allow() {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no token"})
		return
	}
	fmt.Println(c.Request.URL.String())
	if c.Request.URL.Path == "/login" && c.Request.Method == http.MethodPost {
		login(c)
		return
	}

	ps := strings.Split(c.Request.URL.Path, "/")
	l := len(ps)
	if l < 3 {
		log.Error("path err path =", c.Request.URL.Path)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "path is err"})
		return
	}
	c.Set("service_name", ps[l-2])
	c.Set("method", ps[l-1])

	if !verify(c) {
		return
	}

	transmit(c)
}

func loadExternal() error {
	list, err := data.QueryExternalList("configmgr", "login")
	if err != nil {
		log.Errorf("QueryExternalList err = %v", err)
		return err
	}
	for _, e := range list {
		if e.Status == 0 {
			continue
		}
		var m = make(map[string]*data.External)
		mds := strings.Split(e.Method, ",")
		for _, md := range mds {
			m[md] = e
		}
		d, ok := external.LoadOrStore(e.ServiceName, m)
		if !ok {
			continue
		}
		for k, v := range d.(map[string]*data.External) {
			m[k] = v
		}
		external.Store(e.ServiceName, m)
	}

	return nil
}

func Run(host string, port int) error {
	var cfg Config
	err := client.LoadConfig("gateway.yaml", &cfg, "yaml")
	log.Info("Run rate", cfg.Limit, cfg.Bucket, err)

	err = client.WatchConfig("gateway.yaml", Config{}, "yaml")
	if err != nil {
		log.Errorf("WatchConfig err = %v", err)
		return err
	}

	err = loadExternal()
	if err != nil {
		log.Errorf("loadExternal err = %v", err)
		return err
	}

	engine := gin.Default()
	engine.Use(Proxy)
	err = engine.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Errorf("Run err = %v", err)
		return err
	}
	return nil
}
