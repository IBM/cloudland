package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/IBM/cloudland/web/sca/selfs"
)

var (
	RootCmd = &cobra.Command{
		Use:           "",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	Version = "v1.0.0"
)

func init() {
	RootCmd.Version = Version
	selfs.Version = Version
	viper.SetEnvPrefix("CLADMIN")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("admin.endpoint", "127.0.0.1:50080")
	viper.SetDefault("admin.listen", "127.0.0.1:50080")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(-1)
	}
}
