package Logger

import (
	"Qunexx/AuthService/Configs"
	"fmt"
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
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

	currentDate := time.Now()
	logg := fmt.Sprintf("./Logs/%s.log", currentDate.Format("2006-01-02"))

	file, err := os.OpenFile(logg, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Не удалось открыть файл logs.txt" + err.Error())
	}
	fileEncoder := zapcore.NewJSONEncoder(config.EncoderConfig)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), config.Level)

	logger, err := config.Build()
	if err != nil {
		panic("Ошибка сборки логгера " + err.Error())
	}
	defer logger.Sync()

	sentryCore, err := zapsentry.NewCore(zapsentry.Configuration{
		Level: zapcore.ErrorLevel, // Отправлять в Sentry сообщения уровня Error и выше
	}, zapsentry.NewSentryClientFromClient(sentry.CurrentHub().Client()))
	if err != nil {
		panic("Ошибка создания Zap/Sentry: " + err.Error())
	}

	logger = zap.New(zapcore.NewTee(logger.Core(), sentryCore, fileCore))

	return logger
}
