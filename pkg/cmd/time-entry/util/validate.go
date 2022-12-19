package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lucassabreu/clockify-cli/api"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
)

// GetValidateTimeEntryFn will check if the time entry is valid given the
// workspace parameters
func GetValidateTimeEntryFn(f cmdutil.Factory) Step {
	if f.Config().GetBool(cmdutil.CONF_ALLOW_INCOMPLETE) {
		return skip
	}

	return func(tei TimeEntryDTO) (TimeEntryDTO, error) {
		return tei, validateTimeEntry(tei, f)
	}
}

func validateTimeEntry(te TimeEntryDTO, f cmdutil.Factory) error {
	w, err := f.GetWorkspace()
	if err != nil {
		return err
	}

	if w.Settings.ForceProjects && te.ProjectID == "" {
		return errors.New("workspace requires project")
	}

	if w.Settings.ForceTasks && te.TaskID == "" {
		return errors.New("workspace requires task")
	}

	if w.Settings.ForceDescription && strings.TrimSpace(te.Description) == "" {
		return errors.New("workspace requires description")
	}

	if w.Settings.ForceTags && len(te.TagIDs) == 0 {
		return errors.New("workspace requires at least one tag")
	}

	if te.ProjectID == "" {
		return nil
	}

	c, err := f.Client()
	if err != nil {
		return err
	}

	p, err := c.GetProject(api.GetProjectParam{
		Workspace: te.Workspace,
		ProjectID: te.ProjectID,
	})

	if err != nil {
		return err
	}

	if p.Archived {
		return fmt.Errorf("project %s - %s is archived", p.ID, p.Name)
	}

	return nil
}
