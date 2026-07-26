package main

import (
	"flag"
	"fmt"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/account_mgr/internal/config"
	"github.com/starslipay/account_mgr/internal/server"
	"github.com/starslipay/account_mgr/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/accountmgr.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		account_mgr_pb.RegisterAccountMgrServer(grpcServer, server.NewAccountMgrServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// c2cConsumer := consumer.NewC2CConsumer(ctx)
	// err := c2cConsumer.Start(context.Background())
	// if err != nil {
	// 	fmt.Printf("Failed to start C2C consumer: %v\n", err)
	// } else {
	// 	defer c2cConsumer.Stop()
	// }

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
	// go func() {
	// 	s.Start()
	// }()

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// <-sigChan
}
