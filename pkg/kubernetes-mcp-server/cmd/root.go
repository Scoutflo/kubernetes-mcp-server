package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/mark3labs/mcp-go/server"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/mcp"
	"github.com/scoutflo/kubernetes-mcp-server/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
)

var rootCmd = &cobra.Command{
	Use:   "kubernetes-mcp-server [command] [options]",
	Short: "Kubernetes Model Context Protocol (MCP) server",
	Long: `
Kubernetes Model Context Protocol (MCP) server

  # show this help
  kubernetes-mcp-server -h

  # shows version information
  kubernetes-mcp-server --version

  # start STDIO server
  kubernetes-mcp-server

  # start a SSE server on port 8080
  kubernetes-mcp-server --sse-port 8080

  # start a SSE server on port 8443 with a public HTTPS host of example.com
  kubernetes-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443

  # TODO: add more examples`,
	Run: func(cmd *cobra.Command, args []string) {
		initLogging()
		if viper.GetBool("version") {
			fmt.Println(version.Version)
			return
		}
		mcpServer, err := mcp.NewSever()
		if err != nil {
			panic(err)
		}
		defer mcpServer.Close()

		var sseServer *server.SSEServer
		if ssePort := viper.GetInt("sse-port"); ssePort > 0 {
			sseServer = mcpServer.ServeSse(viper.GetString("sse-base-url"))
			defer func() { _ = sseServer.Shutdown(cmd.Context()) }()
			klog.V(0).Infof("SSE server starting on port %d", ssePort)
			if err := sseServer.Start(fmt.Sprintf(":%d", ssePort)); err != nil {
				klog.Errorf("Failed to start SSE server: %s", err)
				return
			}
		}
		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	rootCmd.Flags().IntP("log-level", "", 0, "Set the log level (from 0 to 9)")
	rootCmd.Flags().IntP("sse-port", "", 0, "Start a SSE server on the specified port")
	rootCmd.Flags().StringP("sse-base-url", "", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func initLogging() {
	logger := textlogger.NewLogger(textlogger.NewConfig(textlogger.Output(os.Stdout)))
	klog.SetLoggerWithOptions(logger)
	flagSet := flag.NewFlagSet("kubernetes-mcp-server", flag.ContinueOnError)
	klog.InitFlags(flagSet)
	if logLevel := viper.GetInt("log-level"); logLevel >= 0 {
		_ = flagSet.Parse([]string{"--v", strconv.Itoa(logLevel)})
	}
}
