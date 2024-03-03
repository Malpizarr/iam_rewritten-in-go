package main

import (
	model "AuditService/Model"
	repository "AuditService/Repository"
	grpcService "AuditService/grpcService"
	pb "AuditService/proto"
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	port string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	viper.AutomaticEnv()

	port = viper.GetString("grpcPort")

	var err error
	dsn := viper.GetString("db")
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.AuditEvent{})
	if err != nil {
		return
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	auditRepo := repository.NewAuditEventRepository(db)
	auditService := grpcService.NewAuditServiceImpl(auditRepo)

	s := grpc.NewServer()
	pb.RegisterAuditServiceServer(s, auditService)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
