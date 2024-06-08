package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zanzhit/opencast_recorder/internal/handler"
	"github.com/zanzhit/opencast_recorder/internal/httpserver"
	"github.com/zanzhit/opencast_recorder/internal/repository"
	"github.com/zanzhit/opencast_recorder/internal/service"
	"github.com/zanzhit/opencast_recorder/internal/service/opencast"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing config: %s", err.Error())
	}

	acl, err := json.Marshal(viper.Get("acl"))
	if err != nil {
		logrus.Fatalf("error acl marshalling: %s", err.Error())
	}

	processing, err := json.Marshal(viper.Get("processing"))
	if err != nil {
		logrus.Fatalf("error processing marshalling: %s", err.Error())
	}

	db, err := repository.NewSQLiteDB("sqlite/sqlite")
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	videoService := opencast.NewOpencastService(
		acl,
		processing,
		viper.GetString("videos_path"),
		viper.GetString("video_service"),
		viper.GetString("login"),
		viper.GetString("password"))

	services := service.NewService(repos, videoService, viper.GetString("videos_path"))
	handlers := handler.NewHandler(services)

	srv := new(httpserver.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("Recorder Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Recorder Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down %s", err.Error())
	}
	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
