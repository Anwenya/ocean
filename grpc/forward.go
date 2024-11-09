package grpc

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/mwitkow/grpc-proxy/proxy"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
)

// 连接到目标 gRPC 服务
func ccFunc(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
	log.Println(fullMethodName)
	backendConn, err := grpc.NewClient("localhost:6334", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to backend service: %v", err)
	}
	return ctx, backendConn, nil
}

func ForwardServerStart() {
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cx := cmux.New(listener)
	grpcListener := cx.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := cx.Match(cmux.HTTP1Fast())

	go func() {
		grpcServer := grpc.NewServer(
			grpc.UnknownServiceHandler(proxy.TransparentHandler(ccFunc)),
		)
		err = grpcServer.Serve(grpcListener)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		gin.SetMode(gin.ReleaseMode)
		ginServer := gin.Default()
		ginServer.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "hello world")
		})
		err = http.Serve(httpListener, ginServer)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	if err := cx.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
