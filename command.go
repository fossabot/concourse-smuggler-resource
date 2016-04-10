package smuggler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SmugglerCommand struct {
	lastCommand              *exec.Cmd
	lastCommandCombinedOuput string
}

func NewSmugglerCommand() *SmugglerCommand {
	return &SmugglerCommand{}
}

func (command *SmugglerCommand) LastCommandCombinedOuput() string {
	return command.lastCommandCombinedOuput
}

func (command *SmugglerCommand) Run(commandDefinition CommandDefinition, params map[string]string, outputDir string) error {
	path := commandDefinition.Path
	args := commandDefinition.Args

	params_env := make([]string, len(params)+1)
	params_env = append(params_env, fmt.Sprintf("SMUGGLER_OUTPUT_DIR=%s", outputDir))
	for k, v := range params {
		params_env = append(params_env, fmt.Sprintf("SMUGGLER_%s=%s", k, v))
	}
	params_env = append(params_env, os.Environ()...)

	log.Printf("[INFO] Running '%s %s' with env:\n\t",
		path, strings.Join(args, " "), strings.Join(params_env, "\n\t"))

	command.lastCommand = exec.Command(path, args...)
	command.lastCommand.Env = params_env
	output, err := command.lastCommand.CombinedOutput()
	if err != nil {
		return err
	}
	command.lastCommandCombinedOuput = string(output)
	log.Printf("[INFO] Output '%s'", command.LastCommandCombinedOuput())
	return nil
}

func (command *SmugglerCommand) RunCheck(request CheckRequest) (CheckResponse, error) {
	var response = CheckResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return response, errors.New(message)
	}

	smugglerConfig := request.Source.SmugglerConfig
	if smugglerConfig.CheckCommand.IsDefined() {
		outputDir, err := ioutil.TempDir("", "smuggler-run")
		if err != nil {
			return response, err
		}
		defer os.RemoveAll(outputDir)

		err = command.Run(smugglerConfig.CheckCommand, request.Source.ExtraParams, outputDir)
		if err != nil {
			return response, err
		}

		if versionLines, err := readAndTrimAllLines(filepath.Join(outputDir, "versions")); err != nil {
			return response, err
		} else {
			for _, l := range versionLines {
				response = append(response, Version{VersionID: l})
			}
		}

	}
	return response, nil
}

func (command *SmugglerCommand) RunIn(request InRequest) (InResponse, error) {
	var response = InResponse{}

	if ok, message := request.Source.IsValid(); !ok {
		return InResponse{}, errors.New(message)
	}

	smugglerConfig := request.Source.SmugglerConfig
	if smugglerConfig.InCommand.IsDefined() {
		outputDir, err := ioutil.TempDir("", "smuggler-run")
		if err != nil {
			return response, err
		}
		defer os.RemoveAll(outputDir)

		params := make(map[string]string, len(request.Source.ExtraParams)+len(request.Params))
		for k, v := range request.Source.ExtraParams {
			params[k] = v
		}
		for k, v := range request.Params {
			params[k] = v
		}

		err = command.Run(smugglerConfig.InCommand, params, outputDir)
		if err != nil {
			return response, err
		}

		// We always use the same version that we get in the request
		response.Version = request.Version

		if metadataLines, err := readAndTrimAllLines(filepath.Join(outputDir, "metadata")); err != nil {
			return response, err
		} else {
			for _, l := range metadataLines {
				s := strings.SplitN(l, "=", 2)
				k, v := "", ""
				k = strings.Trim(s[0], " \t")
				if len(s) > 1 {
					v = strings.Trim(s[1], " \t")
				}
				response.Metadata = append(response.Metadata, MetadataPair{Name: k, Value: v})
			}
		}
	}
	return response, nil
}

func readAndTrimAllLines(filename string) ([]string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []string{}, nil
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}
	fileLines := strings.Split(string(content), "\n")
	validLines := make([]string, 0, len(fileLines))
	for _, l := range fileLines {
		trimmedLine := strings.Trim(l, "\t ")
		if trimmedLine != "" {
			validLines = append(validLines, trimmedLine)
		}
	}
	return validLines, nil
}
