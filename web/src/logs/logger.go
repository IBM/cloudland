/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package logs

import (
	"fmt"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

var (
	defaultFields = logrus.Fields{}
	defaultLogger = logrus.New()
)

type SpanLogger struct {
	opentracing.Span
	entry *logrus.Entry
}

func (sp *SpanLogger) logToSpan(args ...interface{}) {
	if sp.Span == nil {
		return
	}
	kvs := make([]interface{}, 0, len(sp.entry.Data)*2+2)
	for k, v := range sp.entry.Data {
		kvs = append(kvs, k, v)
	}
	msg := fmt.Sprint(args...)
	if msg != "" {
		kvs = append(kvs, "msg", msg)
	}
	sp.LogKV(kvs...)
}

func (l *SpanLogger) log(f func(...interface{}), args ...interface{}) {
	l.logToSpan(args...)
	l.entry.Time = time.Now().UTC()
	f(args...)
}

func (l *SpanLogger) logf(f func(string, ...interface{}), format string, args ...interface{}) {
	l.logToSpan(fmt.Sprintf(format, args...))
	l.entry.Time = time.Now().UTC()
	f(format, args...)
}

// Info is for informational messages
func (l *SpanLogger) Info(args ...interface{}) {
	l.log(l.entry.Info, args...)
}

// Debug is for stuff relevant only to a developer
func (l *SpanLogger) Debug(args ...interface{}) {
	l.log(l.entry.Debug, args...)
}

// Warning really ought not be used
func (l *SpanLogger) Warning(args ...interface{}) {
	l.log(l.entry.Warning, args...)
}

// Error is for when things don't go according to plan.
func (l *SpanLogger) Error(args ...interface{}) {
	if l.Span != nil {
		l.SetTag("error", true)
	}
	l.log(l.entry.Error, args...)
}

func (l *SpanLogger) Debugf(format string, args ...interface{}) {
	l.logf(l.entry.Debugf, format, args...)
}

func (l *SpanLogger) Infof(format string, args ...interface{}) {
	l.logf(l.entry.Infof, format, args...)
}

func (l *SpanLogger) Warningf(format string, args ...interface{}) {
	l.logf(l.entry.Warningf, format, args...)
}

func (l *SpanLogger) Errorf(format string, args ...interface{}) {
	if l.Span != nil {
		l.SetTag("error", true)
	}
	l.logf(l.entry.Errorf, format, args...)
}

// WithField allows adding a single key/value pair.
func (l *SpanLogger) WithField(key string, value interface{}) *SpanLogger {
	return &SpanLogger{Span: l.Span, entry: l.entry.WithField(key, value)}
}

// WithFields allows adding multiple kv pairs in a struct.
func (l *SpanLogger) WithFields(arg logrus.Fields) *SpanLogger {
	return &SpanLogger{Span: l.Span, entry: l.entry.WithFields(arg)}
}

// WithError allows shorthand adding of an error field.
func (l *SpanLogger) WithError(err error) *SpanLogger {
	return &SpanLogger{Span: l.Span, entry: l.entry.WithError(err)}
}

// Finish
func (l *SpanLogger) Finish() {
	if sp := l.Span; sp != nil {
		sp.Finish()
	}
}
