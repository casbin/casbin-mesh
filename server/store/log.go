// Copyright 2023 The Casbin Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type raftLogger struct {
	logger *zap.Logger
}

func (r *raftLogger) Trace(msg string, args ...interface{}) {
}

func (r *raftLogger) Debug(msg string, args ...interface{}) {
	r.logger.Debug(msg, argsToFields(args...)...)
}

func (r *raftLogger) Info(msg string, args ...interface{}) {
	r.logger.Info(msg, argsToFields(args...)...)
}

func (r *raftLogger) Warn(msg string, args ...interface{}) {
	r.logger.Warn(msg, argsToFields(args...)...)
}

func (r *raftLogger) Error(msg string, args ...interface{}) {
	r.logger.Error(msg, argsToFields(args...)...)
}

func (r *raftLogger) IsTrace() bool {
	return false
}

func (r *raftLogger) IsDebug() bool {
	return false

}

func (r raftLogger) IsInfo() bool {
	return false

}

func (r *raftLogger) IsWarn() bool {
	return false

}

func (r *raftLogger) IsError() bool {
	return false

}

func (r *raftLogger) With(args ...interface{}) hclog.Logger {
	return &raftLogger{logger: r.logger.With(argsToFields(args...)...)}
}

func (r *raftLogger) Named(name string) hclog.Logger {
	return &raftLogger{logger: r.logger.Named(name)}
}

func (r *raftLogger) ResetNamed(name string) hclog.Logger {
	return r
}

func (r *raftLogger) SetLevel(level hclog.Level) {
}

func (r *raftLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return zap.NewStdLog(r.logger)
}

func (r *raftLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return io.Discard
}

func newRaftLogger(logger *zap.Logger) hclog.Logger {
	return &raftLogger{logger: logger.WithOptions(zap.AddCallerSkip(1))}
}

func argsToFields(args ...interface{}) []zapcore.Field {
	var fields []zapcore.Field

	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if ok {
			fields = append(fields, zap.Any(k, args[i+1]))
		}
	}

	return fields
}
