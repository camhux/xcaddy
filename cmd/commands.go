package xcaddycmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/caddyserver/xcaddy"
	"github.com/caddyserver/xcaddy/internal/utils"
	"github.com/spf13/cobra"
)

func init() {
	buildCommand.Flags().StringArray("with", []string{}, "caddy modules package path to include in the build")
	buildCommand.Flags().String("output", "", "change the output file name")
	buildCommand.Flags().StringArray("replace", []string{}, "like --with but for Go modules")
	buildCommand.Flags().StringArray("embed", []string{}, "embeds directories into the built Caddy executable to use with the `embedded` file-system")
}

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Prints xcaddy version",
	RunE: func(cm *cobra.Command, args []string) error {
		fmt.Println(xcaddyVersion())
		return nil
	},
}

var buildCommand = &cobra.Command{
	Use: `build [<caddy_version>]
    [--output <file>]
    [--with <module[@version][=replacement]>...]
    [--replace <module[@version]=replacement>...]
    [--embed <[alias]:path/to/dir>...]`,
	Long: `
<caddy_version> is the core Caddy version to build; defaults to CADDY_VERSION env variable or latest.
This can be the keyword latest, which will use the latest stable tag, or any git ref such as:

A tag like v2.0.1
A branch like master
A commit like a58f240d3ecbb59285303746406cab50217f8d24

Flags: 
 --output changes the output file.

 --with can be used multiple times to add plugins by specifying the Go module name and optionally its version, similar to go get. Module name is required, but specific version and/or local replacement are optional.

 --replace is like --with, but does not add a blank import to the code; it only writes a replace directive to go.mod, which is useful when developing on Caddy's dependencies (ones that are not Caddy modules). Try this if you got an error when using --with, like cannot find module providing package.

 --embed can be used multiple times to embed directories into the built Caddy executable. The directory can be prefixed with a custom alias and a colon : to use it with the root directive and sub-directive. Example: xcaddy build --embed foo:./sites/foo --embed bar:./sites/bar (This allows you to serve 2 sites from 2 different embedded directories, which are referenced by aliases, from a single Caddy executable).
`,
	Short: "Compile custom caddy binaries",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var output string
		var plugins []xcaddy.Dependency
		var replacements []xcaddy.Replace
		var embedDir []string
		var argCaddyVersion string
		if len(args) > 0 {
			argCaddyVersion = args[0]
		}
		withArgs, err := cmd.Flags().GetStringArray("with")
		if err != nil {
			return fmt.Errorf("unable to parse --with arguments: %s", err.Error())
		}

		replaceArgs, err := cmd.Flags().GetStringArray("replace")
		if err != nil {
			return fmt.Errorf("unable to parse --replace arguments: %s", err.Error())
		}
		for _, withArg := range withArgs {
			mod, ver, _, err := splitWith(withArg)
			if err != nil {
				return err
			}
			mod = strings.TrimSuffix(mod, "/") // easy to accidentally leave a trailing slash if pasting from a URL, but is invalid for Go modules
			plugins = append(plugins, xcaddy.Dependency{
				PackagePath: mod,
				Version:     ver,
			})
		}

		for _, withArg := range replaceArgs {
			mod, ver, repl, err := splitWith(withArg)
			if err != nil {
				return err
			}
			// adjust relative replacements in current working directory since our temporary module is in a different directory
			if strings.HasPrefix(repl, ".") {
				repl, err = filepath.Abs(repl)
				if err != nil {
					log.Fatalf("[FATAL] %v", err)
				}
				log.Printf("[INFO] Resolved relative replacement %s to %s", withArg, repl)
			}
			replacements = append(replacements, xcaddy.NewReplace(xcaddy.Dependency{PackagePath: mod, Version: ver}.String(), repl))
		}

		output, err = cmd.Flags().GetString("output")
		if err != nil {
			return fmt.Errorf("unable to parse --output arguments: %s", err.Error())
		}

		embedDir, err = cmd.Flags().GetStringArray("embed")
		if err != nil {
			return fmt.Errorf("unable to parse --embed arguments: %s", err.Error())
		}
		// prefer caddy version from command line argument over env var
		if argCaddyVersion != "" {
			caddyVersion = argCaddyVersion
		}

		// ensure an output file is always specified
		if output == "" {
			output = getCaddyOutputFile()
		}

		// perform the build
		builder := xcaddy.Builder{
			Compile: xcaddy.Compile{
				Cgo: os.Getenv("CGO_ENABLED") == "1",
			},
			CaddyVersion: caddyVersion,
			Plugins:      plugins,
			Replacements: replacements,
			RaceDetector: raceDetector,
			SkipBuild:    skipBuild,
			SkipCleanup:  skipCleanup,
			Debug:        buildDebugOutput,
			BuildFlags:   buildFlags,
			ModFlags:     modFlags,
		}
		for _, md := range embedDir {
			builder.EmbedDirs = append(builder.EmbedDirs, xcaddy.NewEmbedDir(md))
		}
		err = builder.Build(cmd.Root().Context(), output)
		if err != nil {
			log.Fatalf("[FATAL] %v", err)
		}

		// done if we're skipping the build
		if builder.SkipBuild {
			return nil
		}

		// if requested, run setcap to allow binding to low ports
		err = setcapIfRequested(output)
		if err != nil {
			return err
		}

		// prove the build is working by printing the version
		if runtime.GOOS == utils.GetGOOS() && runtime.GOARCH == utils.GetGOARCH() {
			if !filepath.IsAbs(output) {
				output = "." + string(filepath.Separator) + output
			}
			fmt.Println()
			fmt.Printf("%s version\n", output)
			cmd := exec.Command(output, "version")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				log.Fatalf("[FATAL] %v", err)
			}
		}

		return nil
	},
}
