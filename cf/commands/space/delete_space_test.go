package space_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/organizations/organizationsfakes"
	"github.com/cloudfoundry/cli/cf/api/spaces/spacesfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-space command", func() {
	var (
		ui                  *testterm.FakeUI
		space               models.Space
		config              coreconfig.Repository
		spaceRepo           *spacesfakes.FakeSpaceRepository
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-space").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-space", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		config = testconfig.NewRepositoryWithDefaults()

		space = models.Space{SpaceFields: models.SpaceFields{
			Name: "space-to-delete",
			GUID: "space-to-delete-guid",
		}}

		space_fields := make([]models.SpaceFields, 1)

		space_fields[0] = models.SpaceFields{
			Name: "space-to-delete",
			GUID: "space-to-delete-guid",
		}

		org := models.Organization{
			OrganizationFields: models.OrganizationFields{
				Name: "other-org",
				GUID: "other-org-guid"},
			Spaces: space_fields,
		}

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
		spaceReq := new(requirementsfakes.FakeSpaceRequirement)
		spaceReq.GetSpaceReturns(space)
		orgReq := new(requirementsfakes.FakeOrganizationRequirement)
		orgReq.GetOrganizationReturns(org)
		requirementsFactory.NewSpaceRequirementReturns(spaceReq)
		requirementsFactory.NewOrganizationRequirementReturns(orgReq)
	})

	Describe("requirements", func() {
		BeforeEach(func() {
			ui.Inputs = []string{"y"}
		})
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("my-space")).To(BeFalse())
		})

		It("fails when not targeting a space", func() {
			targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

			Expect(runCommand("my-space")).To(BeFalse())
		})
	})

	It("deletes a space, given its name", func() {
		ui.Inputs = []string{"yes"}
		runCommand("space-to-delete")

		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the space space-to-delete"}))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Deleting space", "space-to-delete", "my-org", "my-user"},
			[]string{"OK"},
		))
		Expect(spaceRepo.DeleteArgsForCall(0)).To(Equal("space-to-delete-guid"))
		Expect(config.HasSpace()).To(Equal(true))
	})

	It("deletes a space in a specified org when the -o flag is given", func() {
		ui.Inputs = []string{"yes"}

		runCommand("-o", "other-org", "space-to-delete")

		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the space space-to-delete"}))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Deleting space", "space-to-delete", "other-org", "my-user"},
			[]string{"OK"},
		))

		// We were on this line - we wanted to figure out how to either test that the correct org
		// guid was used in the space guid lookup call, or fake out the thing that returns
		// a space guid to change which one it gives back depending on what org guide you call it with

		Expect(spaceRepo.DeleteArgsForCall(0)).To(Equal("space-to-delete-guid"))
		Expect(orgRepo.FindByNameCallCount()).To(Equal(1))
		Expect(config.HasSpace()).To(Equal(true))
	})

	It("does not prompt when the -f flag is given", func() {
		runCommand("-f", "space-to-delete")

		Expect(ui.Prompts).To(BeEmpty())
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Deleting", "space-to-delete"},
			[]string{"OK"},
		))
		Expect(spaceRepo.DeleteArgsForCall(0)).To(Equal("space-to-delete-guid"))
	})

	It("clears the space from the config, when deleting the space currently targeted", func() {
		config.SetSpaceFields(space.SpaceFields)
		runCommand("-f", "space-to-delete")

		Expect(config.HasSpace()).To(Equal(false))
	})

	It("clears the space from the config, when deleting the space currently targeted even if space name is case insensitive", func() {
		config.SetSpaceFields(space.SpaceFields)
		runCommand("-f", "Space-To-Delete")

		Expect(config.HasSpace()).To(Equal(false))
	})
})
