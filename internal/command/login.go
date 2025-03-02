// Copyright (c) 2020 tickstep.
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
package command

import (
	"fmt"
	"github.com/tickstep/aliyunpan-api/aliyunpan"
	"github.com/tickstep/aliyunpan/cmder"
	"github.com/tickstep/aliyunpan/internal/config"
	_ "github.com/tickstep/library-go/requester"
	"github.com/urfave/cli"
)


func CmdLogin() cli.Command {
	return cli.Command{
		Name:  "login",
		Usage: "登录阿里云盘账号",
		Description: `
	示例:
		aliyunpan login
		aliyunpan login -RefreshToken=8B12CBBCE89CA8DFC3445985B63B511B5E7EC7...

	常规登录:
		按提示一步一步来即可.
`,
		Category: "阿里云盘账号",
		Before:   cmder.ReloadConfigFunc, // 每次进行登录动作的时候需要调用刷新配置
		After:    cmder.SaveConfigFunc, // 登录完成需要调用保存配置
		Action: func(c *cli.Context) error {
			webToken := aliyunpan.WebLoginToken{}
			refreshToken := ""
			var err error
			refreshToken, webToken, err = RunLogin(c.String("RefreshToken"))
			if err != nil {
				fmt.Println(err)
				return err
			}

			cloudUser, err := config.SetupUserByCookie(&webToken)
			if cloudUser == nil {
				fmt.Println("登录失败: ", err)
				return nil
			}
			cloudUser.RefreshToken = refreshToken
			config.Config.SetActiveUser(cloudUser)
			fmt.Println("阿里云盘登录成功: ", cloudUser.Nickname)
			return nil
		},
		// 命令的附加options参数说明，使用 help login 命令即可查看
		Flags: []cli.Flag{
			// aliyunpan login -RefreshToken=8B12CBBCE89CA8DFC3445985B63B511B5E7EC7...
			cli.StringFlag{
				Name:  "RefreshToken",
				Usage: "使用 RefreshToken Cookie来登录帐号",
			},
		},
	}
}

func CmdLogout() cli.Command {
	return cli.Command{
		Name:        "logout",
		Usage:       "退出阿里帐号",
		Description: "退出当前登录的帐号",
		Category:    "阿里云盘账号",
		Before:      cmder.ReloadConfigFunc,
		After:       cmder.SaveConfigFunc,
		Action: func(c *cli.Context) error {
			if config.Config.NumLogins() == 0 {
				fmt.Println("未设置任何帐号, 不能退出")
				return nil
			}

			var (
				confirm    string
				activeUser = config.Config.ActiveUser()
			)

			if !c.Bool("y") {
				fmt.Printf("确认退出当前帐号: %s ? (y/n) > ", activeUser.Nickname)
				_, err := fmt.Scanln(&confirm)
				if err != nil || (confirm != "y" && confirm != "Y") {
					return err
				}
			}

			deletedUser, err := config.Config.DeleteUser(activeUser.UserId)
			if err != nil {
				fmt.Printf("退出用户 %s, 失败, 错误: %s\n", activeUser.Nickname, err)
			}

			fmt.Printf("退出用户成功: %s\n", deletedUser.Nickname)
			return nil
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "y",
				Usage: "确认退出帐号",
			},
		},
	}
}

func RunLogin(refreshToken string) (refreshTokenStr string, webToken aliyunpan.WebLoginToken, error error) {
	return cmder.DoLoginHelper(refreshToken)
}
