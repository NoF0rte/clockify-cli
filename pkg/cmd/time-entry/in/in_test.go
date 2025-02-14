package in_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/lucassabreu/clockify-cli/api"
	"github.com/lucassabreu/clockify-cli/api/dto"
	"github.com/lucassabreu/clockify-cli/internal/mocks"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/in"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
	"github.com/lucassabreu/clockify-cli/pkg/timehlp"
	"github.com/stretchr/testify/assert"
)

var w = dto.Workspace{ID: "w"}

func TestNewCmdIn_ShouldBeBothBillableAndNotBillable(t *testing.T) {
	f := mocks.NewMockFactory(t)

	f.EXPECT().GetUserID().Return("u", nil)
	f.EXPECT().GetWorkspaceID().Return(w.ID, nil)

	f.EXPECT().Config().Return(&mocks.SimpleConfig{})

	c := mocks.NewMockClient(t)
	f.EXPECT().Client().Return(c, nil)

	called := false
	cmd := in.NewCmdIn(f, func(
		_ dto.TimeEntryImpl, _ io.Writer, _ util.OutputFlags) error {
		called = true
		return nil
	})

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	out := bytes.NewBufferString("")
	cmd.SetOut(out)
	cmd.SetErr(out)

	cmd.SetArgs([]string{"--billable", "--not-billable"})
	_, err := cmd.ExecuteC()

	if assert.Error(t, err) {
		assert.False(t, called)
		flagErr := &cmdutil.FlagError{}
		assert.ErrorAs(t, err, &flagErr)
		return
	}

	t.Fatal("should've failed")
}

func TestNewCmdIn_ShouldNotSetBillable_WhenNotAsked(t *testing.T) {
	bTrue := true
	bFalse := false

	tts := []struct {
		name  string
		args  []string
		param api.CreateTimeEntryParam
	}{
		{
			name: "should be nil",
			args: []string{"-s=08:00"},
			param: api.CreateTimeEntryParam{
				Workspace: w.ID,
				Start:     timehlp.Today().Add(8 * time.Hour),
				Billable:  nil,
			},
		},
		{
			name: "should be billable",
			args: []string{"-s=08:00", "--billable"},
			param: api.CreateTimeEntryParam{
				Workspace: w.ID,
				Start:     timehlp.Today().Add(8 * time.Hour),
				Billable:  &bTrue,
			},
		},
		{
			name: "should not be billable",
			args: []string{"-s=08:00", "--not-billable"},
			param: api.CreateTimeEntryParam{
				Workspace: w.ID,
				Start:     timehlp.Today().Add(8 * time.Hour),
				Billable:  &bFalse,
			},
		},
	}

	for i := range tts {
		tt := &tts[i]

		t.Run(tt.name, func(t *testing.T) {
			f := mocks.NewMockFactory(t)

			f.EXPECT().GetUserID().Return("u", nil)
			f.EXPECT().GetWorkspace().Return(w, nil)
			f.EXPECT().GetWorkspaceID().Return(w.ID, nil)

			f.EXPECT().Config().Return(&mocks.SimpleConfig{
				AllowNameForID: true,
			})

			c := mocks.NewMockClient(t)
			f.EXPECT().Client().Return(c, nil)

			c.EXPECT().GetTimeEntryInProgress(api.GetTimeEntryInProgressParam{
				Workspace: w.ID,
				UserID:    "u",
			}).
				Return(nil, nil)

			c.EXPECT().Out(api.OutParam{
				Workspace: w.ID,
				UserID:    "u",
				End:       tt.param.Start,
			}).Return(api.ErrorNotFound)

			c.EXPECT().CreateTimeEntry(tt.param).
				Return(dto.TimeEntryImpl{ID: "te"}, nil)

			called := false
			cmd := in.NewCmdIn(f, func(
				_ dto.TimeEntryImpl, _ io.Writer, _ util.OutputFlags) error {
				called = true
				return nil
			})

			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			out := bytes.NewBufferString("")
			cmd.SetOut(out)
			cmd.SetErr(out)

			cmd.SetArgs(append(tt.args, "-q"))
			_, err := cmd.ExecuteC()

			if assert.NoError(t, err) {
				t.Cleanup(func() {
					assert.True(t, called)
				})
				return
			}

			t.Fatalf("err: %s", err)
		})
	}

}
