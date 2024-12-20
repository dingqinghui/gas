/**
 * @Author: dingQingHui
 * @Description:
 * @File: config
 * @Version: 1.0.0
 * @Date: 2024/9/23 11:14
 */

package zlog

import (
	"github.com/dingqinghui/gas/api"
	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"path/filepath"
)

func initConfig(node api.INode) *config {
	c := new(config)
	viper := node.GetViper().Sub("log")

	c.level = zapcore.Level(viper.GetInt("level"))
	path := viper.GetString("path")
	c.path = filepath.Join(path, convertor.ToString(node.GetID()), "log")
	c.printConsole = viper.GetBool("printConsole")
	return c
}

type config struct {
	path         string
	level        zapcore.Level
	printConsole bool
	zapOption    []zap.Option
	writer       io.Writer
}

func (o *config) getWriter() io.Writer {
	if o.writer != nil {
		return o.writer
	}
	return defaultWriter(o.getPath())
}

func (o *config) getPath() string {
	if o.path != "" {
		return o.path
	}
	return "./log"
}

func (o *config) getLevel() zapcore.Level {
	return o.level
}
func (o *config) getZapOption() []zap.Option {
	return o.zapOption
}
func (o *config) getPrintConsole() bool {
	return o.printConsole
}
