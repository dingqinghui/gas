/**
 * @Author: dingQingHui
 * @Description:
 * @File: main
 * @Version: 1.0.0
 * @Date: 2024/12/2 15:17
 */

package main

import (
	"github.com/dingqinghui/gas/examples/nodes/chat"
	"github.com/dingqinghui/gas/examples/nodes/gate"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/urfave/cli"
	"os"
)

var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "config,c",
		Usage: "配置路径",
		Value: "../",
	},
}

func main() {
	defer xerror.PrintCoreDump()
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "chat",
			Flags: Flags,
			Action: func(args *cli.Context) error {
				chat.RunChatNode(args.String("config"))
				return nil
			},
		},
		{
			Name:  "gate",
			Flags: Flags,
			Action: func(args *cli.Context) error {
				gate.RunGateNode(args.String("config"))
				return nil
			},
		},
	}
	app.Run(os.Args)
}
