package flags

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/prysmaticlabs/prysm/cmd"
	"github.com/prysmaticlabs/prysm/testing/require"
	"github.com/urfave/cli/v2"
)

func TestLoadFlagsFromConfig_PreProcessing_Web3signer(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	context := cli.NewContext(&app, set, nil)

	pubkey1 := "0xbd36226746676565cd40141a7f0fe1445b9a3fbeb222288b226392c4b230ed0b"
	pubkey2 := "0xbd36226746676565cd40141a7f0fe1445b9a3fbeb222288b226392c4b230ed0a"

	require.NoError(t, os.WriteFile("flags_test.yaml", []byte(fmt.Sprintf("%s:\n - %s\n - %s\n", Web3SignerPublicValidatorKeysFlag.Name,
		pubkey1,
		pubkey2)), 0666))

	require.NoError(t, set.Parse([]string{"test-command", "--" + cmd.ConfigFileFlag.Name, "flags_test.yaml"}))
	command := &cli.Command{
		Name: "test-command",
		Flags: cmd.WrapFlags([]cli.Flag{
			&cli.StringFlag{
				Name: cmd.ConfigFileFlag.Name,
			},
			&cli.StringSliceFlag{
				Name: Web3SignerPublicValidatorKeysFlag.Name,
			},
		}),
		Before: func(cliCtx *cli.Context) error {
			return cmd.LoadFlagsFromConfig(cliCtx, cliCtx.Command.Flags)
		},
		Action: func(cliCtx *cli.Context) error {
			require.Equal(t, strings.Join([]string{pubkey1, pubkey2}, ","),
				strings.Join(cliCtx.StringSlice(Web3SignerPublicValidatorKeysFlag.Name), ","))
			return nil
		},
	}
	require.NoError(t, command.Run(context))
	require.NoError(t, os.Remove("flags_test.yaml"))
}
