// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"fmt"
	"log"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	log         *log.Logger
	EnableDebug bool
}

func NewLogger(name string) *Logger {
	logFile := fmt.Sprintf(config.LogFileTemplate, name)

	return &Logger{
		log: log.New(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    1,
			MaxBackups: 2,
		}, "", log.LstdFlags),
	}
}

func (l *Logger) Info(format string, v ...any) {
	l.log.Println("INFO", fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(format string, v ...any) {
	l.log.Println("WARN", fmt.Sprintf(format, v...))
}

func (l *Logger) Error(format string, v ...any) {
	l.log.Println("ERROR", fmt.Sprintf(format, v...))
}

func (l *Logger) Debug(format string, v ...any) {
	if !l.EnableDebug {
		return
	}
	l.log.Println("DEBUG", fmt.Sprintf(format, v...))
}
