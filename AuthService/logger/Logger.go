package Logger

import (
	"Qunexx/AuthService/Configs"
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func getLogLevel() zapcore.Level {
	Configs.InitEnvConfig()
	levelString := Configs.EnvConfigs.LoggerLevel
	var level zapcore.Level

	switch levelString {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "DPANIC":
		level = zapcore.DPanicLevel
	case "PANIC":
		level = zapcore.PanicLevel
	case "FATAL":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel // Установка уровня по умолчанию, если переменная не задана
	}

	return level
}

func SetupLogger() *zap.Logger {

	if err := sentry.Init(sentry.ClientOptions{
		//Todo Перенести в .env
		Dsn: "https://255124ba32a05531dca58bf92d2cbd9d@o4506949345869824.ingest.us.sentry.io/4506949348098048",
	}); err != nil {
		panic("Sentry initialization failed: " + err.Error())
	}
	defer sentry.Flush(2 * time.Second)

	logLevel := getLogLevel()

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)

	logger, _ := config.Build()
	defer logger.Sync()

	core, err := zapsentry.NewCore(zapsentry.Configuration{
		Level: zapcore.ErrorLevel, // Отправлять в Sentry сообщения уровня Error и выше
	}, zapsentry.NewSentryClientFromClient(sentry.CurrentHub().Client()))
	if err != nil {
		panic("Failed to create zapsentry core: " + err.Error())
	}

	// Создание нового логгера с добавлением Sentry core
	logger = zap.New(zapcore.NewTee(logger.Core(), core))

	return logger
}
