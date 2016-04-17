package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"

	"github.com/redfactorlabs/concourse-smuggler-resource/helpers/utils"
)

var pipeline_yml = Fixture("../../fixtures/pipeline.yml")
var pipeline = NewPipeline(pipeline_yml)
var err error

var _ = Describe("smuggler commands", func() {
	var (
		session *gexec.Session
		logFile *os.File

		commandPath        string
		dataDir            string
		expectedExitStatus int
		request            *ResourceRequest
	)

	BeforeEach(func() {
		expectedExitStatus = 0
		dataDir = ""
	})

	JustBeforeEach(func() {
		var err error
		var command *exec.Cmd

		RegisterFailHandler(Fail)

		stdin := &bytes.Buffer{}
		err = json.NewEncoder(stdin).Encode(request)
		Ω(err).ShouldNot(HaveOccurred())

		if dataDir == "" {
			command = exec.Command(commandPath)
		} else {
			command = exec.Command(commandPath, dataDir)
		}
		command.Stdin = stdin

		// Point log file to a temporary location
		logFile, err = ioutil.TempFile("", "smuggler.log")
		Ω(err).ShouldNot(HaveOccurred())
		command.Env = append(os.Environ(), fmt.Sprintf("SMUGGLER_LOG=%s", logFile.Name()))

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Expect(session.ExitCode()).To(Equal(expectedExitStatus))

	})

	AfterEach(func() {
		stat, err := logFile.Stat()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(stat.Size()).Should(BeNumerically(">", 0))
		os.Remove(logFile.Name())
	})

	Context("when given a complex definition", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				commandPath = checkPath

				commandPath, request = prepareCommandCheck("complex_command")
			})
			It("outputs a valid json with a version", func() {
				var response []interface{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs := JsonStringToInterfaceList([]string{"1.2.3", "1.2.4"})
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("Command Start"))
				Ω(stderr).Should(ContainSubstring("Command End"))
				Ω(stderr).Should(ContainSubstring("param1=test"))
				Ω(stderr).Should(ContainSubstring("param2=true"))
				Ω(stderr).Should(ContainSubstring("param3=123"))
			})
		})
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandIn("complex_command")
			})
			Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests(&session))
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandOut("complex_command")
			})
			Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests(&session))
		})
	})

	Context("when given a dummy command", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				commandPath, request = prepareCommandCheck("dummy_command")
			})

			It("returns empty version list", func() {
				var response []interface{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEmpty())
			})
		})
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandIn("dummy_command")
			})
			It("returns empty response", func() {
				var response ResourceResponse
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.IsEmpty()).Should(BeTrue())
			})
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				commandPath, dataDir, request = prepareCommandOut("dummy_command")
			})
			It("returns empty response", func() {
				var response ResourceResponse
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response.IsEmpty()).Should(BeTrue())
			})
		})
	})

	Context("when given a command which fails", func() {
		Context("for the 'check' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, request = prepareCommandCheck("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
		Context("for the 'in' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, dataDir, request = prepareCommandIn("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
		Context("for the 'out' command", func() {
			BeforeEach(func() {
				expectedExitStatus = 2
				commandPath, dataDir, request = prepareCommandOut("fail_command")
			})

			It("returns an error", func() {
				Ω(session.Err).Should(gbytes.Say("error running command"))
			})
		})
	})

	Context("when there is local config file 'smuggler.yml' that is empty", func() {
		BeforeEach(func() {
			err := utils.Copy("../../fixtures/empty_smuggler.yml",
				filepath.Join(filepath.Dir(checkPath), "smuggler.yml"))
			Ω(err).ShouldNot(HaveOccurred())
		})
		Context("when running 'check' with a dummy definition", func() {
			BeforeEach(func() {
				commandPath, request = prepareCommandCheck("dummy_command")
			})

			It("returns empty version list", func() {
				var response []interface{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(response).Should(BeEmpty())
			})
		})

		Context("when running 'check' with a complex_command definition", func() {
			BeforeEach(func() {
				commandPath = checkPath
				commandPath, request = prepareCommandCheck("complex_command")
			})
			It("outputs a valid json with a version", func() {
				var response []interface{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs := JsonStringToInterfaceList([]string{"1.2.3", "1.2.4"})
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("Command Start"))
				Ω(stderr).Should(ContainSubstring("Command End"))
				Ω(stderr).Should(ContainSubstring("param1=test"))
				Ω(stderr).Should(ContainSubstring("param2=true"))
				Ω(stderr).Should(ContainSubstring("param3=123"))
			})
		})

	})

	Context("when there is local config file 'smuggler.yml' with config", func() {
		BeforeEach(func() {
			err := utils.Copy("../../fixtures/full_smuggler.yml",
				filepath.Join(filepath.Dir(checkPath), "smuggler.yml"))
			Ω(err).ShouldNot(HaveOccurred())
		})
		Context("when running 'check' with a empty command definition", func() {
			BeforeEach(func() {
				commandPath, request = prepareCommandCheck("dummy_command")
			})

			It("returns versions of the config file", func() {
				var response []interface{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs := JsonStringToInterfaceList([]string{"4.5.6", "4.5.7"})
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output from the command definition", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("config_param1=param_in_config"))
				Ω(stderr).Should(ContainSubstring("param1=undef"))
				Ω(stderr).Should(ContainSubstring("from config file"))
				Ω(stderr).ShouldNot(ContainSubstring("Command Start"))
			})
		})

		Context("when running 'check' with a complex command definition", func() {
			BeforeEach(func() {
				commandPath, request = prepareCommandCheck("complex_command")
			})

			It("returns versions of the definition", func() {
				var response []interface{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())
				vs := JsonStringToInterfaceList([]string{"1.2.3", "1.2.4"})
				Ω(response).Should(BeEquivalentTo(vs))
			})

			It("outputs the commands output from the definition", func() {
				stderr := session.Err.Contents()

				Ω(stderr).Should(ContainSubstring("Command Start"))
				Ω(stderr).Should(ContainSubstring("Command End"))
				Ω(stderr).Should(ContainSubstring("param1=test"))
				Ω(stderr).Should(ContainSubstring("param2=true"))
				Ω(stderr).Should(ContainSubstring("param3=123"))
				Ω(stderr).ShouldNot(ContainSubstring("from config file"))
			})
		})
	})

})

func getRequest(t RequestType, resourceName string) *ResourceRequest {
	requestJson, err := pipeline.JsonRequest(t, resourceName, "a_job", "1.2.3")
	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(GinkgoWriter, "%s\n", requestJson)

	request, err := NewResourceRequest(t, requestJson)
	Ω(err).ShouldNot(HaveOccurred())

	return request
}

func prepareCommandCheck(resourceName string) (string, *ResourceRequest) {
	commandPath := checkPath

	request := getRequest(CheckType, resourceName)

	return commandPath, request
}

func prepareCommandIn(resourceName string) (string, string, *ResourceRequest) {
	commandPath := inPath

	tmpPath, err := ioutil.TempDir("", "in_command")
	Ω(err).ShouldNot(HaveOccurred())
	dataDir := filepath.Join(tmpPath, "destination")

	request := getRequest(InType, resourceName)

	return commandPath, dataDir, request
}

func prepareCommandOut(resourceName string) (string, string, *ResourceRequest) {
	commandPath := outPath

	tmpPath, err := ioutil.TempDir("", "in_command")
	Ω(err).ShouldNot(HaveOccurred())
	dataDir := filepath.Join(tmpPath, "destination")

	request := getRequest(OutType, resourceName)

	return commandPath, dataDir, request
}

func InOutCommonSmugglerTests(session **gexec.Session) func() {
	return func() {
		It("outputs a valid json with a version", func() {
			var response ResourceResponse
			err := json.Unmarshal((*session).Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			v := JsonStringToInterface("1.2.3")
			Ω(response.Version).Should(Equal(v))
		})
		It("outputs a valid json with a version", func() {
			var response ResourceResponse
			err := json.Unmarshal((*session).Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
			expectedMetadata := []MetadataPair{
				MetadataPair{Name: "value1", Value: "something quite long"},
				MetadataPair{Name: "value_2", Value: "2"},
			}
			Ω(response.Metadata).Should(Equal(expectedMetadata))
		})
		It("outputs the commands output", func() {
			stderr := (*session).Err.Contents()

			Ω(stderr).Should(ContainSubstring("Command Start"))
			Ω(stderr).Should(ContainSubstring("Command End"))
			Ω(stderr).Should(ContainSubstring("param1=test"))
			Ω(stderr).Should(ContainSubstring("param2=true"))
			Ω(stderr).Should(ContainSubstring("param3=123"))
			Ω(stderr).Should(ContainSubstring("param4=val4"))
			Ω(stderr).Should(ContainSubstring("param5=something with spaces"))
		})
	}
}
