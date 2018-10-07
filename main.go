package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

func main() {
	cmd := NewRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		log.Fatal("Failed to execute command")
	}
}

// NewRootCmd represents the base command when called without any subcommands
func NewRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inject",
		Short: "",
		Long:  ``,
	}

	out := cmd.OutOrStdout()

	cmd.AddCommand(NewUpgradeCommand(out))

	return cmd
}

type upgradeCmd struct {
	injector    string
	release     string
	chart       string
	dryRun      bool
	debug       bool
	valueFiles  []string
	values      []string
	install     bool
	namespace   string
	kubeContext string

	tls     bool
	tlsCert string
	tlsKey  string

	out io.Writer
}

// NewUpgradeCommand represents the upgrade command
func NewUpgradeCommand(out io.Writer) *cobra.Command {
	u := &upgradeCmd{out: out}

	cmd := &cobra.Command{
		Use:   "upgrade [RELEASE] [CHART]",
		Short: "upgrade a release including inject (default injector: linkerd)",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("requires two arguments")
			}
			if u.injector == "helm" {
				return errors.New("why you do this to me")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			release := args[0]
			chart := args[1]

			skip := false

			tempDir, err := copyToTempDir(chart)
			if err != nil {
				skip = true
				fmt.Fprintf(os.Stderr, err.Error())
			}
			fileOptions := fileOptions{
				basePath: tempDir,
				subPath:  "templates",
				fileType: "yaml",
			}
			files, err := getFilesToActOn(fileOptions)
			if err != nil {
				skip = true
			}
			if !skip {
				templateOptions := templateOptions{
					files:       files,
					chart:       tempDir,
					name:        release,
					namespace:   u.namespace,
					values:      u.values,
					valuesFiles: u.valueFiles,
				}
				if err := template(templateOptions); err != nil {
					skip = true
				}
			}
			if !skip {
				injectOptions := injectOptions{
					injector: u.injector,
					files:    files,
				}
				if err := inject(injectOptions); err != nil {
					skip = true
					fmt.Fprintf(os.Stderr, err.Error())
				}
			}
			if !skip {
				upgradeOptions := upgradeOptions{
					chart:       tempDir,
					name:        release,
					values:      u.values,
					valuesFiles: u.valueFiles,
					namespace:   u.namespace,
					kubeContext: u.kubeContext,
					install:     u.install,
					dryRun:      u.dryRun,
					debug:       u.debug,
					tls:         u.tls,
					tlsCert:     u.tlsCert,
					tlsKey:      u.tlsKey,
				}
				if err := upgrade(upgradeOptions); err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
				}
			}

			os.RemoveAll(tempDir)
		},
	}
	f := cmd.Flags()

	f.StringVar(&u.injector, "injector", "linkerd", "injector to use (must be pre-installed)")

	f.StringArrayVarP(&u.valueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL(can specify multiple)")
	f.StringArrayVar(&u.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringVar(&u.namespace, "namespace", "", "namespace to install the release into (only used if --install is set). Defaults to the current kube config namespace")
	f.StringVar(&u.kubeContext, "kube-context", "", "name of the kubeconfig context to use")

	f.BoolVarP(&u.install, "install", "i", false, "if a release by this name doesn't already exist, run an install")
	f.BoolVar(&u.dryRun, "dry-run", false, "simulate an upgrade")
	f.BoolVar(&u.debug, "debug", false, "enable verbose output")

	f.BoolVar(&u.tls, "tls", false, "enable TLS for request")
	f.StringVar(&u.tlsCert, "tls-cert", "", "path to TLS certificate file (default: $HELM_HOME/cert.pem)")
	f.StringVar(&u.tlsKey, "tls-key", "", "path to TLS key file (default: $HELM_HOME/key.pem)")

	return cmd
}

// copyToTempDir checks if the path is local or a repo (in this order) and copies it to a temp directory
// It will perform a `helm fetch` if required
func copyToTempDir(path string) (string, error) {
	tempDir := mkRandomDir(os.TempDir())
	exists, err := exists(path)
	if err != nil {
		return "", err
	}
	if !exists {
		command := fmt.Sprintf("helm fetch %s --untar -d %s", path, tempDir)
		Exec(command)
		files, err := ioutil.ReadDir(tempDir)
		if err != nil {
			return "", err
		}
		if len(files) != 1 {
			return "", fmt.Errorf("%d additional files found in temp direcotry. This is very strange", len(files)-1)
		}
		tempDir = filepath.Join(tempDir, files[0].Name())
	} else {
		err = copy.Copy(path, tempDir)
		if err != nil {
			return "", err
		}
	}
	return tempDir, nil
}

type fileOptions struct {
	basePath string
	subPath  string
	fileType string
}

func getFilesToActOn(options fileOptions) ([]string, error) {
	var files []string
	dir := filepath.Join(options.basePath, options.subPath)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, options.fileType) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

type templateOptions struct {
	files       []string
	chart       string
	name        string
	values      []string
	valuesFiles []string
	namespace   string
}

func template(o templateOptions) error {
	var additionalFlags string
	additionalFlags += createFlagChain("set", o.values)
	defaultValuesPath := filepath.Join(o.chart, "values.yaml")
	exists, err := exists(defaultValuesPath)
	if err != nil {
		return err
	}
	if exists {
		additionalFlags += createFlagChain("f", []string{defaultValuesPath})
	}
	additionalFlags += createFlagChain("f", o.valuesFiles)
	if o.namespace != "" {
		additionalFlags += createFlagChain("namespace", []string{o.namespace})
	}

	for _, file := range o.files {
		command := fmt.Sprintf("helm template --debug=false %s --name %s -x %s%s", o.chart, o.name, file, additionalFlags)
		output := Exec(command)
		if err := ioutil.WriteFile(file, output, 0644); err != nil {
			return err
		}
	}

	return nil
}

type injectOptions struct {
	injector string
	files    []string
}

func inject(o injectOptions) error {
	for _, file := range o.files {
		command := fmt.Sprintf("%s inject %s", o.injector, file)
		output := Exec(command)
		if o.injector == "linkerd" {
			output = removeSummary(output)
		}
		if err := ioutil.WriteFile(file, output, 0644); err != nil {
			return err
		}
	}

	return nil
}

func removeSummary(input []byte) []byte {
	lines := strings.Split(string(input), "\n")
	lastLine := 0
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "---") {
			lastLine = i
			break
		}
	}
	return []byte(strings.Join(lines[:lastLine], "\n"))
}

type upgradeOptions struct {
	chart       string
	name        string
	values      []string
	valuesFiles []string
	namespace   string
	kubeContext string
	install     bool
	dryRun      bool
	debug       bool
	tls         bool
	tlsCert     string
	tlsKey      string
}

func upgrade(o upgradeOptions) error {
	additionalFlags := ""
	additionalFlags = additionalFlags + createFlagChain("set", o.values)
	additionalFlags = additionalFlags + createFlagChain("f", o.valuesFiles)
	if o.namespace != "" {
		additionalFlags = additionalFlags + createFlagChain("namespace", []string{o.namespace})
	}
	if o.kubeContext != "" {
		additionalFlags = additionalFlags + createFlagChain("kube-context", []string{o.kubeContext})
	}
	if o.install {
		additionalFlags = additionalFlags + createFlagChain("i", []string{""})
	}
	if o.dryRun {
		additionalFlags = additionalFlags + createFlagChain("dry-run", []string{""})
	}
	if o.debug {
		additionalFlags = additionalFlags + createFlagChain("debug", []string{""})
	}
	if o.tls {
		additionalFlags = additionalFlags + createFlagChain("tls", []string{""})
	}
	if o.tlsCert != "" {
		additionalFlags = additionalFlags + createFlagChain("debug", []string{o.tlsCert})
	}
	if o.tlsKey != "" {
		additionalFlags = additionalFlags + createFlagChain("debug", []string{o.tlsKey})
	}

	command := fmt.Sprintf("helm upgrade %s %s%s", o.name, o.chart, additionalFlags)
	output := Exec(command)
	fmt.Println(string(output))

	return nil
}

// Exec takes a command as a string and executes it
func Exec(cmd string) []byte {
	args := strings.Split(cmd, " ")
	binary := args[0]
	_, err := exec.LookPath(binary)
	if err != nil {
		log.Fatal(err)
	}

	output, err := exec.Command(binary, args[1:]...).CombinedOutput()
	if err != nil {
		log.Fatal(string(output))
	}
	return output
}

// MkRandomDir creates a new directory with a random name made of numbers
func mkRandomDir(basepath string) string {
	r := strconv.Itoa((rand.New(rand.NewSource(time.Now().UnixNano()))).Int())
	path := filepath.Join(basepath, r)
	os.Mkdir(path, 0755)

	return path
}

func createFlagChain(flag string, input []string) string {
	chain := ""
	dashes := "--"
	if len(flag) == 1 {
		dashes = "-"
	}

	for _, i := range input {
		if i != "" {
			i = " " + i
		}
		chain = fmt.Sprintf("%s %s%s%s", chain, dashes, flag, i)
	}

	return chain
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
