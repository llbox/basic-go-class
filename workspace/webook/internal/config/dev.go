//go:build !k8s

package config

var Config = WeBookConfig{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:30001)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:30002",
	},
}
