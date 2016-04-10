package command

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/redfactorlabs/concourse-smuggler-resource"
)

type SmugglerCommand struct {
	lastCommand              *exec.Cmd
	lastCommandCombinedOuput string
}

func (command *SmugglerCommand) LastCommandCombinedOuput() string {
	return command.lastCommandCombinedOuput
}

func (command *SmugglerCommand) Run(commandDefinition smuggler.CommandDefinition, params map[string]string) error {
	path := commandDefinition.Path
	args := commandDefinition.Args

	params_env := make([]string, len(params))
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
