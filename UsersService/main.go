package main

import (
	"UsersService/Controllers"
	"UsersService/Middleware"
	"UsersService/Repositories"
	"UsersService/Service"
	"UsersService/graph"
	grpcService "UsersService/grpcService"
	"UsersService/model"
	_ "UsersService/model"
	pb "UsersService/proto"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
)

var (
	db       *gorm.DB
	portgrpc string

	graphqlPort string
)

func init() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	viper.AutomaticEnv()

	dsn := viper.GetString("db")
	portgrpc = viper.GetString("grpcPort")
	graphqlPort = viper.GetString("graphqlPort")
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error al conectar con la base de datos: %v", err)
	}

	if err := db.AutoMigrate(&model.GORMUser{}, &model.GORMRole{}, &model.GORMOAuthProvider{}); err != nil {
		log.Fatalf("Error al migrar modelos: %v", err)
	}

}

func main() {
	userRepo := Repositories.NewGormUserRepository(db)

	roleRepo := &Repositories.GormRoleRepository{DB: db}

	userService := Service.NewUserService(userRepo, roleRepo)
	twoFA := Service.NewTwoFactorAuthenticationService()
	jwtService := Service.NewTokenValidationService()

	lis, err := net.Listen("tcp", portgrpc)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	userGrpcService := grpcService.NewUserServiceImpl(userService, twoFA)
	grpcServer := grpc.NewServer()

	pb.RegisterUserServiceServer(grpcServer, userGrpcService)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	log.Println("gRPC server started at port " + portgrpc)

	resolver := Controllers.NewResolver(userService, jwtService)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/graphql", Middleware.AuthMiddleware(srv))

	log.Printf("Conectarse a http://localhost%s/ para el playground de GraphQL", graphqlPort)
	if err := http.ListenAndServe(graphqlPort, nil); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}
