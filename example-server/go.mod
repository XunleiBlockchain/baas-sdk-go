module server

go 1.13

replace (
	github.com/XunleiBlockchain/baas-sdk-go => ../
	github.com/tjfoc/gmsm => github.com/bcscb8/gmsm v0.0.0-20191220070229-b97b35b41ab6
)

require (
	github.com/Terry-Mao/goconf v0.0.0-20161115082538-13cb73d70c44
	github.com/XunleiBlockchain/baas-sdk-go v0.0.0-00010101000000-000000000000
	github.com/binacsgo/log v0.0.0-20200827012301-4f49b8c3150e
	go.uber.org/zap v1.16.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	xorm.io/core v0.7.3 // indirect
)
