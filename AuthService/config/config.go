package config

import (
	repository "AuthService/Repository"
	"AuthService/data"
	"AuthService/grpc"
	"AuthService/service"
	"AuthService/util"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var (
	DB                 *gorm.DB
	UserService        *service.UserService
	EmailService       *service.EmailService
	TokenServiceClient *util.TokenServiceClient
	AuditClient        *grpc.AuditClient
)

func init() {
	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	viper.AutomaticEnv()

	// Configurar conexión a la base de datos
	dsn := viper.GetString("db")
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	err = DB.AutoMigrate(&data.VerificationToken{})
	if err != nil {
		log.Fatal(err)
	}

	// Crear instancias de servicios y clientes
	userClient, err := grpc.NewUserClient("localhost", 9091)
	VerificationTokenRepository := repository.NewVerificationTokenRepository(DB)
	EmailService = service.NewEmailService(
		viper.GetString("FROM"),
		viper.GetString("EMAIL_PASSWORD"),
		viper.GetString("SMTP_HOST"),
		viper.GetString("SMTP_PORT"),
	)
	UserService = service.NewUserService(*EmailService, *VerificationTokenRepository, *userClient)
	AuditClient, err = grpc.NewAuditClient("localhost", 9090)
	TokenServiceClient = util.NewTokenServiceClient()
}
