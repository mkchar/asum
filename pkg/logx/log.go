package logx

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func Set() {
	lvl, err := zerolog.ParseLevel("debug")
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.CallerSkipFrameCount = 3

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

}

func Info(msg string) {
	logger.Info().Msg(msg)
}

func Infof(format string, v ...interface{}) {
	logger.Info().Msgf(format, v...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	ev := logger.Info()
	addFields(ev, keysAndValues...)
	ev.Msg(msg)
}

func Error(msg string) {
	logger.Error().Msg(msg)
}

func Errorf(format string, v ...interface{}) {
	logger.Error().Msgf(format, v...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	ev := logger.Error()
	addFields(ev, keysAndValues...)
	ev.Msg(msg)
}

func Debug(msg string) {
	logger.Debug().Msg(msg)
}

func Debugf(format string, v ...interface{}) {
	logger.Debug().Msgf(format, v...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	ev := logger.Debug()
	addFields(ev, keysAndValues...)
	ev.Msg(msg)
}

func Fatal(msg string) {
	logger.Fatal().Msg(msg)
}

// 内部辅助函数
func addFields(e *zerolog.Event, keysAndValues ...interface{}) {
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if !ok {
				continue
			}
			val := keysAndValues[i+1]
			e.Interface(key, val)
		}
	}
}
