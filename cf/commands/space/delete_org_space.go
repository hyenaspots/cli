package space

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteOrgSpace struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceRepo spaces.SpaceRepository
	orgRepo   organizations.OrganizationRepository
}

func init() {
	commandregistry.Register(&DeleteOrgSpace{})
}

func (cmd *DeleteOrgSpace) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "delete-org-space",
		Description: T("Delete a space within a specific org"),
		Usage: []string{
			T("CF_NAME delete-org-space ORG SPACE"),
		},
	}
}

func (cmd *DeleteOrgSpace) Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) ([]requirements.Requirement, error) {
	if len(context.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires two arguments.\n\n") + commandregistry.Commands.CommandUsage("delete-org-space"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(context.Args()), 2)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *DeleteOrgSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()

	//commandDep := commandregistry.Commands.FindCommand("delete-org-space")
	//commandDep = commandDep.SetDependency(deps, false)

	return cmd
}

func (cmd *DeleteOrgSpace) Execute(c flags.FlagContext) error {
	orgName := c.Args()[0]
	spaceName := c.Args()[1]
	orgGUID := ""
	spaceGUID := ""

	cmd.ui.Say(T("Deleting space!"))

	org, err := cmd.orgRepo.FindByName(orgName)
	if err != nil {
		return err
	}

	orgGUID = org.GUID

	space, err := cmd.spaceRepo.FindByNameInOrg(spaceName, orgGUID)
	if err != nil {
		return err
	}

	spaceGUID = space.GUID

	err = cmd.spaceRepo.Delete(spaceGUID)
	if err != nil {
		return err
	}
	return nil
}
