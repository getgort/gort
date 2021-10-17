/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/client"
	"github.com/getgort/gort/data"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// cogctl bundle install --help
// Usage: cogctl bundle install [OPTIONS] BUNDLE_OR_PATH [VERSION]
//
//   Install a bundle.
//
//   Bundles may be installed from either a file (i.e., the `config.yaml` file
//   of a bundle), or from Operable's Warehouse bundle registry
//   (https://warehouse.operable.io).
//
//   When installing from a file, you may either give the path to the file, as
//   in:
//
//       cogctl bundle install /path/to/my/bundle/config.yaml
//
//   or you may give the path as `-`, in which case standard input is used:
//
//       cat config.yaml | cogctl bundle install -
//
//   When installing from the bundle registry, you should instead provide the
//   name of the bundle, as well as an optional version to install. No version
//   means the latest will be installed.
//
//       cogctl bundle install cfn
//
//       cogctl bundle install cfn 0.5.13
//
// Options:
//   -e, --enable               Automatically enable a bundle after installing?
//                              [default: False]
//   -f, --force                Install even if a bundle with the same version is
//                              already installed. Applies only to bundles
//                              installed from a file, and not from the Warehouse
//                              bundle registry. Use this to shorten iteration
//                              cycles in bundle development.  [default: False]
//   -r, --relay-group TEXT     Relay group to assign the bundle to. Can be
//                              specified multiple times.
//   -t, --templates DIRECTORY  Path to templates directory. Template bodies will
//                              be inserted into the bundle configuration prior
//                              to uploading. This makes it easier to manage
//                              complex templates.
//   --help                     Show this message and exit.

const (
	bundleInstallUse   = "install"
	bundleInstallShort = "Install a bundle"
	bundleInstallLong  = `Install a bundle from a bundle file.

When using this command, you must provide the path to the file, as follows:

  gort bundle install /path/to/my/bundle/config.yaml

You may also give the path as ` + "`-`" + `, in which case standard input is used:

  cat config.yaml | gort bundle install -

Finally, if the bundle lives in git, the latest version of the bundle can be
installed as follows:

  gort bundle install git@example.com/repo.git

A specific tagged version can be installed as follows:

  gort bundle install git@example.com/repo.git@v1.2.3

For bundles in sub-directories, a double-slash syntax can be used to indicate
the subdirectory where the bundle can be found:

  gort bundle install git@example.com/repo.git//bundles/example

If a bundle file is not specified, it will default to "bundle.yml".
`
	bundleInstallUsage = `Usage:
  gort bundle install [flags] config_path

Flags:
  -h, --help			Show this message and exit
  -e, --enable			Automatically enable a bundle after installing?
						[default: False]
  -f, --force			Install even if a bundle with the same version is
						already installed. Applies only to bundles
						installed from a file, and not from the Warehouse
						bundle registry. Use this to shorten iteration
						cycles in bundle development.  [default: False]

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagBundleInstallEnable bool
	flagBundleInstallForce  bool
)

// GetBundleInstallCmd is a command
func GetBundleInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleInstallUse,
		Short: bundleInstallShort,
		Long:  bundleInstallLong,
		RunE:  bundleInstallCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(bundleInstallUsage)
	cmd.Flags().BoolVarP(&flagBundleInstallEnable, "enable", "e", false, "Automatically enable a bundle after installing")
	cmd.Flags().BoolVarP(&flagBundleInstallForce, "force", "f", false, "Install even if a bundle with the same version is already installed.")

	return cmd
}

func bundleInstallCmd(cmd *cobra.Command, args []string) error {
	bundlefile := args[0]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	var bundle data.Bundle

	switch {
	case bundlefile == "-":
		bundle, err = bundles.LoadBundle(os.Stdin)

	case strings.HasSuffix(bundlefile, ".git"):
		bundle, err = loadBundleRepo(bundlefile, "", "")

	case strings.Contains(bundlefile, ".git//"):
		ss := strings.Split(bundlefile, ".git//")
		bundle, err = loadBundleRepo(ss[0]+".git", "", ss[1])

	case strings.Contains(bundlefile, ".git@"):
		ss := strings.Split(bundlefile, ".git@")
		repo := ss[0] + ".git"

		ss = strings.Split(ss[1], "//")
		tag := ss[0]

		var path string
		if len(ss) == 2 {
			path = ss[1]
		}

		bundle, err = loadBundleRepo(repo, tag, path)

	default:
		bundle, err = loadBundleFile(bundlefile)
	}

	if err != nil {
		return err
	}

	// Check for existing instances of this bundle, allowing forced replacement.
	exists, err := c.BundleExists(bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if exists {
		if !flagBundleInstallForce {
			return fmt.Errorf("bundle %q already exists at version %s. Use --force to force install", bundle.Name, bundle.Version)
		}
		err = c.BundleUninstall(bundle.Name, bundle.Version)
		if err != nil {
			return err
		}
	}

	err = c.BundleInstall(bundle)
	if err != nil {
		return err
	}

	if flagBundleInstallEnable {
		err = c.BundleEnable(bundle.Name, bundle.Version)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Bundle %q installed.\n", bundle.Name)

	return nil
}

func loadBundleFile(bundlefile string) (data.Bundle, error) {
	file, err := os.Open(bundlefile)
	if err != nil {
		return data.Bundle{}, err
	}
	defer file.Close()

	return bundles.LoadBundle(file)
}

func loadBundleRepo(repo, tag, path string) (data.Bundle, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "bundle")
	if err != nil {
		return data.Bundle{}, err
	}
	defer os.RemoveAll(dir)

	if tag == "" {
		tag = "latest"
	}

	fmt.Printf("Cloning remote repository: %s@%s//%s\n", repo, tag, path)

	if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
		path += "/bundle.yml"
	}

	fmt.Println("Installing bundle:", path)

	path = fmt.Sprintf("%s/%s", dir, path)

	co := &git.CloneOptions{URL: repo}

	r, err := git.PlainClone(dir, false, co)
	if err != nil {
		return data.Bundle{}, err
	}

	if tag != "latest" {
		ref, err := r.Tag(tag)
		if err != nil {
			return data.Bundle{}, fmt.Errorf("could not install tag %s@%s: %w", repo, tag, err)
		}

		w, err := r.Worktree()
		if err != nil {
			return data.Bundle{}, err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: ref.Hash(),
		})
		if err != nil {
			return data.Bundle{}, err
		}
	}

	return loadBundleFile(path)
}
