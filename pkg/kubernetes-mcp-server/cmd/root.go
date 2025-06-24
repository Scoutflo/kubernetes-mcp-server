package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

  # start a SSE server on port 8080 (supports multiple concurrent clients)
  kubernetes-mcp-server --sse-port 8080

  # start a SSE server on port 8443 with a public HTTPS host of example.com
  kubernetes-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443

Note: 
- SSE mode supports multiple concurrent client connections
- Health checks are available on port 8082

  # TODO: add more examples`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println(version.Version)
			return
		}
		initLogging()
		mcpServer, err := mcp.NewSever()
		if err != nil {
			panic(err)
		}

		ssePort := viper.GetInt("sse-port")
		if ssePort <= 0 {
			fmt.Println("Error: --sse-port is required and must be greater than 0")
			os.Exit(1)
		}

		sseServer := mcpServer.ServeSse(viper.GetString("sse-base-url"))

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		errChan := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("SSE server panic recovered: %v", r)
					errChan <- fmt.Errorf("SSE server panic: %v", r)
				}
			}()

			klog.V(0).Infof("SSE server starting on port %d", ssePort)
			if err := sseServer.Start(fmt.Sprintf(":%d", ssePort)); err != nil {
				errChan <- err
			}
		}()

		klog.V(0).Infof("SSE server running on port %d, supporting multiple concurrent clients", ssePort)
		klog.V(0).Infof("Server health check available on port %d", mcp.HealthPort)

		select {
		case sig := <-sigChan:
			klog.V(0).Infof("Received signal %v, initiating graceful shutdown...", sig)

			klog.V(0).Infof("Waiting 2 seconds for active connections to complete...")
			time.Sleep(2 * time.Second)

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			func() {
				defer func() {
					if r := recover(); r != nil {
						klog.Errorf("Panic during shutdown recovered: %v", r)
					}
				}()

				if err := sseServer.Shutdown(shutdownCtx); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						klog.Warningf("SSE server shutdown timed out, forcing shutdown...")
					} else {
						klog.Errorf("Error during SSE server shutdown: %v", err)
					}
				} else {
					klog.V(0).Infof("SSE server shut down gracefully")
				}
			}()

		case err := <-errChan:
			klog.Errorf("SSE server error: %s", err)
			return
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	rootCmd.Flags().IntP("log-level", "", 2, "Set the log level (from 0 to 9, default 2 to show all tool calls and details)")
	rootCmd.Flags().IntP("sse-port", "", 0, "Start a SSE server on the specified port (required)")
	rootCmd.Flags().StringP("sse-base-url", "", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func initLogging() {
	logLevel := viper.GetInt("log-level")
	if logLevel < 0 {
		logLevel = 2
	}

	logOutput := os.Stdout

	config := textlogger.NewConfig(
		textlogger.Output(logOutput),
		textlogger.Verbosity(logLevel),
	)
	logger := textlogger.NewLogger(config)
	klog.SetLoggerWithOptions(logger)

	flagSet := flag.NewFlagSet("kubernetes-mcp-server", flag.ContinueOnError)
	klog.InitFlags(flagSet)
	if err := flagSet.Parse([]string{"--v", strconv.Itoa(logLevel)}); err != nil {
		fmt.Fprintf(logOutput, "Error parsing log level: %v\n", err)
	}

	klog.V(0).Infof("Logging initialized with level %d", logLevel)
}
