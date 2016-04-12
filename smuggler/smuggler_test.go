package smuggler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"

	. "github.com/redfactorlabs/concourse-smuggler-resource/helpers/test"
	. "github.com/redfactorlabs/concourse-smuggler-resource/smuggler"
)

var manifest = Fixture("pipeline.yml")
var logger = log.New(GinkgoWriter, "smuggler: ", log.Lmicroseconds)

var _ = Describe("Check Command", func() {
	Context("when given a basic config from a structure", func() {
		request := CheckRequest{
			Source: Source{
				Commands: []CommandDefinition{
					CommandDefinition{
						Name: "check",
						Path: "bash",
						Args: []string{"-e", "-c", "echo basic echo test"},
					},
				},
			},
		}

		It("it executes the command successfully and captures the output", func() {
			command := NewSmugglerCommand(logger)
			command.RunCheck(request)
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})

	Context("when given a basic config from a json", func() {
		var command *SmugglerCommand
		var response CheckResponse
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
			request, err := NewCheckRequestFromJson(requestJson)
			Ω(err).ShouldNot(HaveOccurred())
			command = NewSmugglerCommand(logger)
			response, err = command.RunCheck(request)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("it executes the command successfully and captures the output", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("basic echo test"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
	})

	Context("when given a config with empty config from yaml", func() {
		var request CheckRequest
		var command *SmugglerCommand
		var response CheckResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "dummy_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunCheck(request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("does not execute and returns an empty response", func() {
			Ω(command.LastCommand()).Should(BeNil())
			Ω(command.LastCommandCombinedOuput()).Should(BeEmpty())
			Ω(command.LastCommandSuccess()).Should(BeTrue())
			Ω(response).Should(BeEmpty())
		})
	})
	Context("when given a config with a complex script from yaml", func() {
		var request CheckRequest
		var command *SmugglerCommand
		var response CheckResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "complex_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunCheck(request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("executes several lines of the script", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command Start"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command End"))
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
		It("it sets the $MUGGLER_ACTION and $SMUGGLE_COMMAND variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("action=check"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("command=check"))
		})
		It("it can sets the resource extra_params as environment variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
		})
		It("it returns versions as list of strings", func() {
			vs := []Version{Version{VersionID: "1.2.3"}, Version{VersionID: "1.2.4"}}
			Ω(response).Should(BeEquivalentTo(vs))
		})
	})
	Context("when given a command which fails", func() {
		var request CheckRequest
		var command *SmugglerCommand
		var response CheckResponse

		It("captures the exit code", func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "fail_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = CheckRequest{
				Source: *source,
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunCheck(request)
			Ω(err).Should(HaveOccurred())

			Ω(command.LastCommandSuccess()).Should(BeFalse())
			Ω(command.LastCommandExitStatus()).Should(Equal(2))
		})
	})
})

var _ = Describe("In Command", func() {
	Context("when given a config with a complex script from yaml", func() {
		var request InRequest
		var command *SmugglerCommand
		var response InResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "complex_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = InRequest{
				Source:  *source,
				Version: Version{VersionID: "1.2.3"},
				Params: map[string]string{
					"param4": "val4",
					"param5": "something with spaces",
				},
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunIn("/tmp/destination/dir", request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("executes several lines of the script", func() {
			Ω(command.LastCommandSuccess()).Should(BeTrue())
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command Start"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command End"))
		})
		It("it sets the $MUGGLER_ACTION and $SMUGGLE_COMMAND variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("action=in"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("command=in"))
		})
		It("it sets the resource extra_params and 'get' params as environment variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param4=val4"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param5=something with spaces"))
		})

		It("it gets the destination dir", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("destinationDir=/tmp/destination/dir"))
		})

		It("it gets the version ID", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("version=1.2.3"))
		})

		It("it returns metadata as list of strings", func() {
			vs := []MetadataPair{
				MetadataPair{Name: "value1", Value: "something quite long"},
				MetadataPair{Name: "value_2", Value: "2"},
			}
			Ω(response.Metadata).Should(BeEquivalentTo(vs))
		})
	})

	Context("when given a dummy config with empty commands from yaml", func() {
		var request InRequest
		var command *SmugglerCommand
		var response InResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "dummy_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = InRequest{
				Source:  *source,
				Version: Version{VersionID: "1.2.3"},
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunIn("/tmp/destination/dir", request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("does not execute", func() {
			Ω(command.LastCommand()).Should(BeNil())
			Ω(command.LastCommandCombinedOuput()).Should(BeEmpty())
			Ω(command.LastCommandSuccess()).Should(BeTrue())
		})
		It("does not execute and returns the same version", func() {
			Ω(response.Version).Should(Equal(request.Version))
		})
		It("does not execute and returns empty metadata", func() {
			Ω(response.Metadata).Should(BeEmpty())
		})
	})
})

var _ = Describe("Out Command", func() {
	Context("when given a config with a complex script from yaml", func() {
		var request OutRequest
		var command *SmugglerCommand
		var response OutResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "complex_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = OutRequest{
				Source: *source,
				Params: map[string]string{
					"param4": "val4",
					"param5": "something with spaces",
				},
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunOut("/tmp/sources/dir", request)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("executes several lines of the script", func() {
			Ω(command.LastCommandSuccess()).Should(BeTrue())
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command Start"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("Command End"))
		})

		It("it sets the $MUGGLER_ACTION and $SMUGGLE_COMMAND variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("action=out"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("command=out"))
		})
		It("it sets the resource extra_params and 'put' params as environment variables", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param1=test"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param2=true"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param3=123"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param4=val4"))
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("param5=something with spaces"))
		})

		It("it gets the sources dir", func() {
			Ω(command.LastCommandCombinedOuput()).Should(ContainSubstring("sourcesDir=/tmp/sources/dir"))
		})

		It("it return the version ID", func() {
			Ω(response.Version).Should(BeEquivalentTo(Version{VersionID: "1.2.3"}))
		})

		It("it returns metadata as list of strings", func() {
			vs := []MetadataPair{
				MetadataPair{Name: "value1", Value: "something quite long"},
				MetadataPair{Name: "value_2", Value: "2"},
			}
			Ω(response.Metadata).Should(BeEquivalentTo(vs))
		})
	})

	Context("when given a dummy config with empty commands from yaml", func() {
		var request OutRequest
		var command *SmugglerCommand
		var response OutResponse

		BeforeEach(func() {
			source, err := ResourceSourceFromYamlManifest(manifest, "dummy_command")
			Ω(err).ShouldNot(HaveOccurred())
			request = OutRequest{
				Source: *source,
			}
			command = NewSmugglerCommand(logger)
			response, err = command.RunOut("/tmp/sources/dir", request)
			Ω(err).Should(HaveOccurred())
		})
		It("does not execute the command", func() {
			Ω(command.LastCommand()).Should(BeNil())
		})
		It("does not execute and returns empty metadata", func() {
			Ω(response.Metadata).Should(BeEmpty())
		})
	})
})
