package changer_test

import (
	"testing"

	"github.com/cloudfoundry/stack-auditor/cf"

	"github.com/cloudfoundry/stack-auditor/changer"
	"github.com/cloudfoundry/stack-auditor/mocks"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

const (
	AppAName   = "appA"
	AppBName   = "appB"
	AppAGuid   = "appAGuid"
	AppBGuid   = "appBGuid"
	StackAName = "stackA"
	StackBName = "stackB"
	StackAGuid = "stackAGuid"
	StackBGuid = "stackBGuid"
)

func TestUnitChanger(t *testing.T) {
	spec.Run(t, "Changer", testChanger, spec.Report(report.Terminal{}))
}

func testChanger(t *testing.T, when spec.G, it spec.S) {

	var (
		mockCtrl       *gomock.Controller
		mockConnection *mocks.MockCliConnection
		c              changer.Changer
	)

	it.Before(func() {
		RegisterTestingT(t)

		mockCtrl = gomock.NewController(t)
		mockConnection = mocks.SetupMockCliConnection(mockCtrl)

		c = changer.Changer{
			CF: cf.CF{
				Conn: mockConnection,
			},
		}
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	when("running change-stack", func() {
		it("changes the stacks", func() {
			mockConnection.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/"+AppAGuid, "-X", "PUT", `-d={"stack_guid":"`+StackBGuid+`","state":"STOPPED"}`).Return(
				[]string{}, nil)

			mockConnection.EXPECT().CliCommand("start", AppAName)

			result, err := c.ChangeStack(AppAName, StackBName)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("\nApplication appA was successfully changed to Stack stackB"))
		})

		it("returns an error when given an invalid stack", func() {
			_, err := c.ChangeStack(AppAName, "WrongStack")
			Expect(err).To(MatchError("WrongStack is not a valid stack"))
		})

		it("returns an error when given the stack that the app is on", func() {
			_, err := c.ChangeStack(AppAName, StackAName)
			Expect(err).To(MatchError("application is already associated with stack " + StackAName))
		})

		it("returns an error when an app can't be found", func() {
			fakeApp := "appC"
			_, err := c.ChangeStack(fakeApp, StackAName)
			Expect(err).To(MatchError("application could not be found"))
		})
	})
}
