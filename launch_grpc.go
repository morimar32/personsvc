package main

import (
	"context"
	"fmt"
	"net"

	person "personsvc/generated"
	"personsvc/service"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func codeToLevel(code codes.Code) zapcore.Level {
	if code == codes.OK {
		// It is DEBUG
		return zap.DebugLevel
	}
	return grpc_zap.DefaultCodeToLevel(code)
}

func launchGRPC() error {
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}

	opts := []grpc.ServerOption{}

	opts = append(opts, grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(Log, o...),
	))

	lis, err := net.Listen("tcp", grpcaddress)
	if err != nil {
		return err
	}

	db := service.NewPersonRepository(connectionString)
	svc := service.NewPersonService(db, Log)
	s := grpc.NewServer(opts...)
	person.RegisterPersonServer(s, svc)

	if err = db.Ping(context.Background()); err != nil {
		return err
	}

	// Serve gRPC Server
	Log.Info(fmt.Sprintf("Serving gRPC on http://%s", grpcaddress))
	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}
