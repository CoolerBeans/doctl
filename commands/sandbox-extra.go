/*
Copyright 2018 The Doctl Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"errors"
	"fmt"
	"strings"
)

// SandboxExtras adds commands to the 'sandbox' subtree for which the cobra wrappers were autogenerated from
// oclif equivalents and subsequently modified.
func SandboxExtras(cmd *Command) {

	create := CmdBuilder(cmd, RunSandboxExtraCreate, "init <path>", "Initialize a local file system directory for the sandbox",
		`The `+"`"+`doctl sandbox init`+"`"+` command specifies a directory in your file system which will hold functions and
supporting artifacts while you're developing them.  When ready, you can upload these to the cloud for
testing.  Later, after the area is committed to a `+"`"+`git`+"`"+` repository, you can create an app from them.
Type `+"`"+`doctl sandbox status --languages`+"`"+` for a list of supported languages.`,
		Writer)
	AddStringFlag(create, "language", "l", "javascript", "Language for the initial sample code")
	AddBoolFlag(create, "overwrite", "", false, "Clears and reuses an existing directory")

	deploy := CmdBuilder(cmd, RunSandboxExtraDeploy, "deploy <directories>", "Deploy sandbox local assets to the cloud",
		`At any time you can use `+"`"+`doctl sandbox deploy`+"`"+` to upload the contents of a directory in your file system for
testing in the cloud.  The area must be organized in the fashion expected by an App Platform Functions
component.  The `+"`"+`doctl sandbox init`+"`"+` command will create a properly organized directory for you to work in.`,
		Writer)
	AddStringFlag(deploy, "env", "", "", "Path to runtime environment file")
	AddStringFlag(deploy, "build-env", "", "", "Path to build-time environment file")
	AddStringFlag(deploy, "apihost", "", "", "API host to use")
	AddStringFlag(deploy, "auth", "", "", "OpenWhisk auth token to use")
	AddBoolFlag(deploy, "insecure", "", false, "Ignore SSL Certificates")
	AddBoolFlag(deploy, "verbose-build", "", false, "Display build details")
	AddBoolFlag(deploy, "verbose-zip", "", false, "Display start/end of zipping phase for each function")
	AddBoolFlag(deploy, "yarn", "", false, "Use yarn instead of npm for node builds")
	AddStringFlag(deploy, "include", "", "", "Functions and/or packages to include")
	AddStringFlag(deploy, "exclude", "", "", "Functions and/or packages to exclude")
	AddBoolFlag(deploy, "remote-build", "", false, "Run builds remotely")
	AddBoolFlag(deploy, "incremental", "", false, "Deploy only changes since last deploy")

	getMetadata := CmdBuilder(cmd, RunSandboxExtraGetMetadata, "get-metadata <directory>", "Obtain metadata of a sandbox directory",
		`The `+"`"+`doctl sandbox get-metadata`+"`"+` command produces a JSON structure that summarizes the contents of a directory
you have designated for functions development.  This can be useful for feeding into other tools.`,
		Writer)
	AddStringFlag(getMetadata, "env", "", "", "Path to environment file")
	AddStringFlag(getMetadata, "include", "", "", "Functions or packages to include")
	AddStringFlag(getMetadata, "exclude", "", "", "Functions or packages to exclude")

	watch := CmdBuilder(cmd, RunSandboxExtraWatch, "watch <directory>", "Watch a sandbox directory, deploying incrementally on change",
		`Type `+"`"+`doctl sandbox watch <directory>`+"`"+` in a separate terminal window.  It will run until interrupted.
It will watch the directory (which should be one you initialized for sandbox use) and will deploy
the contents to the cloud incrementally as it detects changes.`,
		Writer)
	AddStringFlag(watch, "env", "", "", "Path to runtime environment file")
	AddStringFlag(watch, "build-env", "", "", "Path to build-time environment file")
	AddStringFlag(watch, "apihost", "", "", "API host to use")
	AddStringFlag(watch, "auth", "", "", "OpenWhisk auth token to use")
	AddBoolFlag(watch, "insecure", "", false, "Ignore SSL Certificates")
	AddBoolFlag(watch, "verbose-build", "", false, "Display build details")
	AddBoolFlag(watch, "verbose-zip", "", false, "Display start/end of zipping phase for each function")
	AddBoolFlag(watch, "yarn", "", false, "Use yarn instead of npm for node builds")
	AddStringFlag(watch, "include", "", "", "Functions and/or packages to include")
	AddStringFlag(watch, "exclude", "", "", "Functions and/or packages to exclude")
	AddBoolFlag(watch, "remote-build", "", false, "Run builds remotely")
}

// RunSandboxExtraCreate supports the 'sandbox init' command
func RunSandboxExtraCreate(c *CmdConfig) error {
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	output, err := RunSandboxExec(projectCreate, c, []string{flagOverwrite}, []string{flagLanguage})
	if err != nil {
		// Fix up error message text
		text := err.Error()
		if strings.Contains(text, "already exists") {
			text = strings.Replace(text, "-o", "--overwrite", 1)
			err = errors.New(text)
		}
		return err
	}
	// Special processing for output, since PrintSandboxTextOutput will emit the 'nim' hint which
	// is not quite right for doctl.
	if jsonOutput, ok := output.Entity.(map[string]interface{}); ok {
		if created, ok := jsonOutput["project"].(string); ok {
			fmt.Fprintf(c.Out, `A local sandbox area '%s' was created for you.
You may deploy it by running the command shown on the next line:
  doctl sandbox deploy %s
`, created, created)
			fmt.Fprintln(c.Out)
			return nil
		}
	}
	// Fall back if output is not structured the way we expect
	fmt.Println("Sandbox initialized successfully in the local file system")
	return nil
}

// RunSandboxExtraDeploy supports the 'sandbox deploy' command
func RunSandboxExtraDeploy(c *CmdConfig) error {
	adjustIncludeAndExclude(c)
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	output, err := RunSandboxExec(projectDeploy, c, []string{flagInsecure, flagVerboseBuild, flagVerboseZip, flagYarn, flagRemoteBuild, flagIncremental},
		[]string{flagEnv, flagBuildEnv, flagApihost, flagAuth, flagInclude, flagExclude})
	if err != nil && len(output.Captured) == 0 {
		// Just an error, nothing in 'Captured'
		return err
	}
	// The output from "project/deploy" is not quite right for doctl even with branding, so fix up
	// what is in 'Captured'.  We do this even if there has been an error, because the output of
	// deploy is complex and the transcript is often needed to interpret the error.
	for index, value := range output.Captured {
		if strings.Contains(value, "Deploying project") {
			output.Captured[index] = strings.Replace(value, "Deploying project", "Deployed", 1)
		} else if strings.Contains(value, "Deployed actions") {
			output.Captured[index] = "Deployed functions ('doctl sbx fn get <funcName> --url' for URL):"
		}
	}
	if err == nil {
		// Normal error-free return
		return c.PrintSandboxTextOutput(output)
	}
	// When there is an error but also a transcript, display the transcript before return the error
	// This is "best effort" so we ignore any error returns from the print statement
	fmt.Fprintln(c.Out, strings.Join(output.Captured, "\n"))
	return err
}

// RunSandboxExtraGetMetadata supports the 'sandbox get-metadata' command
func RunSandboxExtraGetMetadata(c *CmdConfig) error {
	adjustIncludeAndExclude(c)
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	output, err := RunSandboxExec(projectGetMetadata, c, []string{flagJSON}, []string{flagEnv, flagInclude, flagExclude})
	if err != nil {
		return err
	}
	return c.PrintSandboxTextOutput(output)
}

// RunSandboxExtraWatch supports 'sandbox watch'
// This is not the usual boiler-plate because the command is intended to be long-running in a separate window
func RunSandboxExtraWatch(c *CmdConfig) error {
	adjustIncludeAndExclude(c)
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	return RunSandboxExecStreaming(projectWatch, c, []string{flagInsecure, flagVerboseBuild, flagVerboseZip, flagYarn, flagRemoteBuild},
		[]string{flagEnv, flagBuildEnv, flagApihost, flagAuth, flagInclude, flagExclude})
}

// adjustIncludeAndExclude deals with the fact that 'web' has special meaning to 'nim'.
// 1.  If the developer has a package called 'web' and wishes to include or exclude it, 'nim' will be confused unless a trailing
// slash is added to indicate that the intent is the package called 'web' and not 'web content.'
// 2.  Since projects may have a non-empty 'web' folder, 'nim' will want to deploy it unless '--exclude web' is provided.
// Note that the developer may already by using '--exclude', so this additional exclusion will be an append to the existing
// value.
func adjustIncludeAndExclude(c *CmdConfig) {
	includes, err := c.Doit.GetString(c.NS, flagInclude)
	if err == nil && includes != "" {
		includes = qualifyWebWithSlash(includes)
		c.Doit.Set(c.NS, flagInclude, includes)
	}
	excludes, err := c.Doit.GetString(c.NS, flagExclude)
	if err == nil && excludes != "" {
		excludes = qualifyWebWithSlash(excludes)
		excludes = excludes + "," + keywordWeb
		c.Doit.Set(c.NS, flagExclude, excludes)
	} else {
		c.Doit.Set(c.NS, flagExclude, keywordWeb)
	}
}

// qualifyWebWithSlash is a subroutine used by adjustIncludeAndExclude.  Given a comma-separated
// list of tokens, if any of those tokens are 'web', change that token to 'web/' and return the
// modified list.
func qualifyWebWithSlash(original string) string {
	tokens := strings.Split(original, ",")
	for i, token := range tokens {
		if token == "web" {
			tokens[i] = "web/"
		}
	}
	return strings.Join(tokens, ",")
}
