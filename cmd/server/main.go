package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/echotalk/echotalk_server/internal/bootstrap"
	"github.com/echotalk/echotalk_server/internal/config"
	"github.com/echotalk/echotalk_server/internal/module/payment"
	"github.com/echotalk/echotalk_server/internal/module/payment/channel"
	"github.com/echotalk/echotalk_server/internal/module/training"
	"github.com/echotalk/echotalk_server/internal/module/user/codesender"
	"github.com/echotalk/echotalk_server/internal/module/user/tokenstore"
	"github.com/echotalk/echotalk_server/internal/pkg/cos"
	"github.com/echotalk/echotalk_server/internal/pkg/jwt"
	"github.com/echotalk/echotalk_server/internal/router"
	"github.com/echotalk/echotalk_server/internal/speech"
	"github.com/echotalk/echotalk_server/internal/speech/iflytek"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := bootstrap.InitLogger(cfg.Log)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	db, err := bootstrap.InitDB(cfg.MySQL)
	if err != nil {
		logger.Fatal("init db failed: " + err.Error())
	}
	if err := bootstrap.AutoMigrate(db); err != nil {
		logger.Fatal("auto migrate failed: " + err.Error())
	}

	rdb, err := bootstrap.InitRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("init redis failed: " + err.Error())
	}

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, cfg.JWT.Issuer)
	speechGW := speech.NewGateway(iflytek.NewISEProvider(cfg.Iflytek), logger)
	payChannel := channel.NewMockChannel()
	codeSender := codesender.NewMockSender()
	codeStore := codesender.NewRedisCodeStore(rdb)
	tokenStore := tokenstore.NewRedisTokenStore(rdb)
	members := payment.NewMembershipChecker(db) // 真实会员门禁（查 memberships 表）

	// 录音存储（可选）：COS 未配置则保持 nil，录音不存，评测照常。
	var audioUploader training.AudioUploader
	if up, err := cos.New(cfg.COS); err != nil {
		logger.Info("COS 未配置，录音不存储: " + err.Error())
	} else {
		audioUploader = up
	}

	engine := router.Setup(router.Deps{
		Config:     cfg,
		DB:         db,
		Logger:     logger,
		JWT:        jwtManager,
		Speech:     speechGW,
		PayChannel: payChannel,
		CodeSender: codeSender,
		CodeStore:  codeStore,
		TokenStore: tokenStore,
		Members:    members,
		Audio:      audioUploader,
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("server starting on " + addr)
	if err := engine.Run(addr); err != nil {
		logger.Fatal("server exited: " + err.Error())
	}
}
