package smuggler_test

import (
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var manifest = Fixture("../fixtures/pipeline.yml")
var logger = log.New(GinkgoWriter, "smuggler: ", log.Lmicroseconds)

var request ResourceRequest
var response ResourceResponse
var command *SmugglerCommand
var fixtureResourceName string
var requestType RequestType
var requestVersion Version
var err error

var _ = Describe("Check Command basic tests", func() {
	Context("when given a basic config from a structure", func() {
		request := ResourceRequest{
			Source: Source{
				Commands: []CommandDefinition{
					CommandDefinition{
						Name: "check",
						Path: "bash",
						Args: []string{"-e", "-c", "echo basic echo test"},
					},
				},
			},
			Type: CheckType,
		}

		It("it executes the command successfully and captures the output", func() {
			command := NewSmugglerCommand(logger)
			command.RunAction("", request)
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})

	Context("when given a basic config from a json", func() {
		requestJson := `{
			"source": {
				"commands": [
					{
						"name": "check",
						"path": "sh",
						"args": [ "-e", "-c", "echo basic echo test" ]
					}
				]
			},
			"version": {}
		}`

		BeforeEach(func() {
			request, err = NewResourceRequestFromJson(requestJson, CheckType)
			Ω(err).ShouldNot(HaveOccurred())
			command = NewSmugglerCommand(logger)
			response, err = command.RunAction("", request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("it executes the command successfully and captures the output", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})
})

var _ = Describe("SmugglerCommand Actions", func() {
	JustBeforeEach(func() {
		request, err = GetResourceRequestFromYamlManifest(requestType, manifest, fixtureResourceName, "a_job")
		Ω(err).ShouldNot(HaveOccurred())
		command = NewSmugglerCommand(logger)
		response, err = command.RunAction("/some/path", request)
	})

	Context("when calling action 'check'", func() {
		BeforeEach(func() {
			requestType = CheckType
		})

		Context("when running CommonSmugglerTests()", CommonSmugglerTests())

		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})
			It("it gets the version id", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("version=1.2.3"))
			})
			It("it returns versions as list of strings", func() {
				vs := []Version{Version{VersionID: "1.2.3"}, Version{VersionID: "1.2.4"}}
				Ω(response.Versions).Should(BeEquivalentTo(vs))
			})
		})
	})

	Context("When calling action 'in'", func() {
		BeforeEach(func() {
			requestType = InType
		})
		Context("when running CommonSmugglerTests()", CommonSmugglerTests())

		Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests())

		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})

			It("it gets the version id", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("version=1.2.3"))
			})
			It("it gets the destination dir", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("destinationDir=/some/path"))
			})
		})
	})
	Context("When calling action 'out'", func() {
		BeforeEach(func() {
			requestType = OutType
		})
		Context("when running CommonSmugglerTests()", CommonSmugglerTests())

		Context("when running InOutCommonSmugglerTests()", InOutCommonSmugglerTests())

		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})
			It("it gets the sources dir", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("sourcesDir=/some/path"))
			})
		})
	})
})

func CommonSmugglerTests() func() {
	return func() {
		Context("when given a config with empty config from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "dummy_command"
			})
			It("executes without errors", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("does not execute and returns an empty response", func() {
				Ω(command.LastCommand()).Should(BeNil())
				Ω(command.LastCommandCombinedOuput()).Should(BeEmpty())
				Ω(command.LastCommandSuccess()).Should(BeTrue())
				Ω(response.IsEmpty()).Should(BeTrue())
			})
		})
		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})

			It("executes without errors", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("executes several lines of the script", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command Start"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command End"))
				Ω(command.LastCommandSuccess()).Should(BeTrue())
			})
			It("it sets the $MUGGLER_ACTION and $SMUGGLE_COMMAND variables", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("action=" + string(requestType)))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("command=" + string(requestType)))
			})
			It("it can sets the resource extra_params as environment variables", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
			})
		})
		Context("when given a command which fails", func() {
			BeforeEach(func() {
				fixtureResourceName = "fail_command"
			})
			It("returns error", func() {
				Ω(err).Should(HaveOccurred())
			})
			It("captures the exit code", func() {
				Ω(command.LastCommandSuccess()).Should(BeFalse())
				Ω(command.LastCommandExitStatus()).Should(Equal(2))
			})
		})
	}
}

func InOutCommonSmugglerTests() func() {
	return func() {
		Context("when given a config with a complex script from yaml", func() {
			BeforeEach(func() {
				fixtureResourceName = "complex_command"
			})
			It("it sets the resource extra_params and 'get' params as environment variables", func() {
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param4=val4"))
				Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param5=something with spaces"))
			})
			It("it returns metadata as list of strings", func() {
				vs := []MetadataPair{
					MetadataPair{Name: "value1", Value: "something quite long"},
					MetadataPair{Name: "value_2", Value: "2"},
				}
				Ω(response.Metadata).Should(BeEquivalentTo(vs))
			})
			It("it returns the version ID", func() {
				Ω(response.Version).Should(BeEquivalentTo(Version{VersionID: "1.2.3"}))
			})
		})
	}
}
