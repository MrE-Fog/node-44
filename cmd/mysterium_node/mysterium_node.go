/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cihub/seelog"
	"github.com/mysterium/node/cmd"
	command_cli "github.com/mysterium/node/cmd/commands/cli"
	command_run "github.com/mysterium/node/cmd/commands/run"
	"github.com/mysterium/node/cmd/commands/version"
	_ "github.com/mysterium/node/logconfig"
	"github.com/mysterium/node/metadata"
	tequilapi_client "github.com/mysterium/node/tequilapi/client"
	"github.com/urfave/cli"
)

func main() {
	err := NewCommand().Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NewCommand function creates new node command
func NewCommand() *cli.App {
	app := cli.NewApp()
	app.Usage = "VPN server and client for Mysterium Network https://mysterium.network/"
	app.Authors = []cli.Author{
		{`The "MysteriumNetwork/node" Authors`, "mysterium-dev@mysterium.network"},
	}
	app.Version = metadata.VersionAsString()
	app.Commands = []cli.Command{
		*versionCommand,
	}
	app.Flags = []cli.Flag{
		tequilapiAddressFlag,
		tequilapiPortFlag,

		licenseWarrantyFlag,
		licenseConditionsFlag,

		openvpnBinaryFlag,
		ipifyUrlFlag,
		locationDatabaseFlag,
		cliFlag,
	}
	app.Action = runMain

	cli.VersionPrinter = func(ctx *cli.Context) {
		versionCommand.Run(ctx)
	}

	return app
}

func runMain(_ *cli.Context) error {
	defer seelog.Flush()

	if options.LicenseWarranty {
		fmt.Println(metadata.LicenseWarranty)
		return nil
	} else if options.LicenseConditions {
		fmt.Println(metadata.LicenseConditions)
		return nil
	} else if options.CLI {
		return runCLI(options)
	} else {
		fmt.Println(versionSummary)
		fmt.Println()

		return runCMD(options)
	}
}

func runCLI(options command_run.CommandOptions) error {
	cmdCli := command_cli.NewCommand(
		filepath.Join(options.Directories.Data, ".cli_history"),
		tequilapi_client.NewClient(options.TequilapiAddress, options.TequilapiPort),
	)
	stop := cmd.HardKiller(cmdCli.Kill)
	cmd.RegisterSignalCallback(stop)

	return cmdCli.Run()
}

func runCMD(options command_run.CommandOptions) error {
	cmdRun := command_run.NewCommand(options)
	stop := cmd.SoftKiller(cmdRun.Kill)
	cmd.RegisterSignalCallback(stop)

	if err := cmdRun.Start(); err != nil {
		return err
	}

	return cmdRun.Wait()
}

var (
	options command_run.CommandOptions

	versionSummary = metadata.VersionAsSummary(metadata.LicenseCopyright(
		"command_run program with '--license.warranty' option",
		"command_run program with '--license.conditions' option",
	))
	versionCommand = version.NewCommand(versionSummary)

	tequilapiAddressFlag = cli.StringFlag{
		Name:        "tequilapi.address",
		Usage:       "IP address of interface to listen for incoming connections",
		Destination: &options.TequilapiAddress,
		Value:       "127.0.0.1",
	}
	tequilapiPortFlag = cli.IntFlag{
		Name:        "tequilapi.port",
		Usage:       "Port for listening incoming api requests",
		Destination: &options.TequilapiPort,
		Value:       4050,
	}

	licenseWarrantyFlag = cli.BoolFlag{
		Name:        "license.warranty",
		Usage:       "Show warranty",
		Destination: &options.LicenseWarranty,
	}
	licenseConditionsFlag = cli.BoolFlag{
		Name:        "license.conditions",
		Usage:       "Show conditions",
		Destination: &options.LicenseConditions,
	}

	openvpnBinaryFlag = cli.StringFlag{
		Name:        "openvpn.binary",
		Usage:       "openvpn binary to use for Open VPN connections",
		Destination: &options.OpenvpnBinary,
		Value:       "openvpn",
	}
	ipifyUrlFlag = cli.StringFlag{
		Name:        "ipify-url",
		Usage:       "Address (URL form) of ipify service",
		Destination: &options.IpifyUrl,
		Value:       "https://api.ipify.org/",
	}
	locationDatabaseFlag = cli.StringFlag{
		Name:        "location.database",
		Usage:       "Service location autodetect database of GeoLite2 format e.g. http://dev.maxmind.com/geoip/geoip2/geolite2/",
		Destination: &options.LocationDatabase,
		Value:       "GeoLite2-Country.mmdb",
	}
	cliFlag = cli.BoolFlag{
		Name:        "cli",
		Usage:       "Run an interactive CLI based Mysterium UI",
		Destination: &options.CLI,
	}
)
