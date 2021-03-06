package user_test

import (
	"cf"
	. "cf/commands/user"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpaceUsersFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSpaceUsers(t, []string{}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSpaceUsers(t, []string{"my-org"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSpaceUsers(t, []string{"my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestSpaceUsersRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{}
	userRepo := &testapi.FakeUserRepository{}
	args := []string{"my-org", "my-space"}

	reqFactory.LoginSuccess = false
	callSpaceUsers(t, args, reqFactory, spaceRepo, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSpaceUsers(t, args, reqFactory, spaceRepo, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, "my-org", reqFactory.OrganizationName)
}

func TestSpaceUsers(t *testing.T) {
	org := cf.Organization{Name: "Org1"}
	space := cf.Space{Name: "Space1"}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameInOrgSpace: space}

	usersByRole := map[string][]cf.User{
		"SPACE MANAGER": []cf.User{
			{Username: "My User 1"},
			{Username: "My User 2"},
		},
		"SPACE DEV": []cf.User{
			{Username: "My User 3"},
		},
	}
	userRepo := &testapi.FakeUserRepository{FindAllInSpaceByRoleUsersByRole: usersByRole}

	ui := callSpaceUsers(t, []string{"my-org", "my-space"}, reqFactory, spaceRepo, userRepo)

	assert.Equal(t, spaceRepo.FindByNameInOrgName, "my-space")
	assert.Equal(t, spaceRepo.FindByNameInOrgOrg, org)

	assert.Contains(t, ui.Outputs[0], "Getting users in org")
	assert.Contains(t, ui.Outputs[0], "Org1")
	assert.Contains(t, ui.Outputs[0], "Space1")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, userRepo.FindAllInSpaceByRoleSpace, space)

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "SPACE MANAGER")
	assert.Contains(t, ui.Outputs[4], "My User 1")
	assert.Contains(t, ui.Outputs[5], "My User 2")

	assert.Contains(t, ui.Outputs[7], "SPACE DEV")
	assert.Contains(t, ui.Outputs[8], "My User 3")
}

func callSpaceUsers(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewSpaceUsers(ui, config, spaceRepo, userRepo)
	ctxt := testcmd.NewContext("space-users", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
