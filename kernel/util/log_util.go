package util

import (
	"fmt"
	"github.com/go-logr/logr"
	. "github.com/sergi/go-diff/diffmatchpatch"
	uberZap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

var L = ctrl.Log.WithName("kernel").WithName("logger")

func DefaultLogger(fileName string, opts ...zap.Opts) logr.Logger {
	defaultOpts := []zap.Opts{
		zap.Encoder(zapcore.NewJSONEncoder(
			zapcore.EncoderConfig{
				TimeKey:       "time",
				LevelKey:      "level",
				NameKey:       "logger",
				CallerKey:     "line",
				MessageKey:    "msg",
				StacktraceKey: "stacktrace",
				LineEnding:    zapcore.DefaultLineEnding,
				EncodeLevel:   zapcore.LowercaseLevelEncoder,
				EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
					enc.AppendString(t.Format("2006-01-02 15:04:05"))
				},
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
				EncodeName:     zapcore.FullNameEncoder,
			})),
		zap.WriteTo(io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    128,
			MaxBackups: 5,
			MaxAge:     7,
			LocalTime:  true,
		})),
		zap.RawZapOpts(uberZap.AddCaller()),
	}

	return zap.New(
		append(defaultOpts, opts...)...,
	)
}

func PrintFingerDiff(observed, desired string) {
	dmp := New()
	diffs := dmp.DiffMain(observed, desired, false)
	std := dmp.DiffPrettyText(diffs)
	fmt.Printf("cr finger diff: %s\n", std)
}
