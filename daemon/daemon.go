package daemon

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/takama/daemon"
)

// CloseFunc 关闭函数接口
type CloseFunc func()

// HandleSystemSignal 处理系统信号
func HandleSystemSignal(sigChan chan os.Signal, cf CloseFunc) {
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	for sig := range sigChan {
		switch sig {
		case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM: //获取到停止信号
			cf()
		case syscall.SIGHUP: //重载配置文件
			//reloadCfg()
		default:
			fmt.Println("signal : ", sig)
		}
	}
}

type namedesc func() (string, string)

// Charge 服务操作
func Charge(name, description string) *Service {
	srv, err := daemon.New(name, description)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	return &Service{srv}
}

// Service .
type Service struct {
	daemon.Daemon
}

// Cmds .
func Cmds(funcs namedesc) []*cobra.Command {
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "安装系统服务，install后面的命令将被追加",
		Run: func(cmd *cobra.Command, args []string) {
			service := Charge(funcs())
			var status string
			var err error
			if len(os.Args) > 2 {
				status, err = service.Install(os.Args[2:]...)
			} else {
				status, err = service.Install()
			}
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(status)
			}
		},
	}
	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "卸载系统服务",
		Run: func(cmd *cobra.Command, args []string) {
			service := Charge(funcs())
			status, err := service.Remove()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(status)
			}
		},
	}
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "启动系统服务",
		Run: func(cmd *cobra.Command, args []string) {
			service := Charge(funcs())
			status, err := service.Start()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(status)
			}
		},
	}
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "停止系统服务",
		Run: func(cmd *cobra.Command, args []string) {
			service := Charge(funcs())
			status, err := service.Stop()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(status)
			}
		},
	}
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "系统服务状态",
		Run: func(cmd *cobra.Command, args []string) {
			service := Charge(funcs())
			status, err := service.Status()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(status)
			}
		},
	}
	return []*cobra.Command{installCmd, removeCmd, startCmd, stopCmd, statusCmd}
}
