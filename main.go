package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"gorm-trace/middleware"
	mysql2 "gorm-trace/mysql"
	redis2 "gorm-trace/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func main() {

	// 初始化trace
	cfg := &jaegerConfig.Configuration{
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const", //固定采样
			Param: 1,       //1=全采样、0=不采样
		},

		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "0.0.0.0:6831",
		},

		ServiceName: "test-trace",
	}

	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err.Error()))
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	app := iris.New()

	// 调用trace 中间
	app.Use(middleware.Trace())

	app.Get("/trace", func(context context.Context) {

		type User struct {
			ID   int64
			Name string
		}
		res := make([]*User, 0)
		// 初始化mysql
		sqlDB := initMysql()
		sqlDB.WithContext(context.Request().Context()).Table("users").Limit(1).Find(&res)

		// 初始化redis
		redisDB := initRedis()
		redisDB.Set(context.Request().Context(), "i-am-key", "i-am-value", time.Hour)

	})

	err = app.Listen(":7787")
	if err != nil {
		panic("init web fail")
	}

}

func initMysql() *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s/%s?charset=utf8mb4&parseTime=True&loc=Local",
			"root:root@tcp(127.0.0.1:3306)", "gongjiayun"), // DSN data source name, parse time is important !!!
		DefaultStringSize:         256,                     // string default length
		SkipInitializeWithVersion: true,                    // auto config according to version
	}), &gorm.Config{})
	if err != nil {
		panic("fail to init mysql: " + err.Error())
	}

	_ = db.Use(&mysql2.OpenTracingPlugin{})

	return db
}

func initRedis() *redis.Client {
	redisDB := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})

	redisDB.AddHook(redis2.TracingHook{})
	return redisDB
}
