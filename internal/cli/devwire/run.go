package devwire

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/config"
)

// RunExplanation executes a Development Experience explain workflow for CLI commands.
func RunExplanation(
	cmd *cobra.Command,
	arg string,
	run func(context.Context, *Handle, string) (*development.DevelopmentExplanation, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, arg)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatExplanation(out), out)
}

// RunPlan executes a Development Experience plan workflow for CLI commands.
func RunPlan(
	cmd *cobra.Command,
	task string,
	run func(context.Context, *Handle, string) (*development.DevelopmentPlan, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, task)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatPlan(out), out)
}

// RunReview executes a Development Experience review workflow for CLI commands.
func RunReview(
	cmd *cobra.Command,
	topic string,
	run func(context.Context, *Handle, string) (*development.DevelopmentReviewGuide, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, topic)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatReviewGuide(out), out)
}

// RunOnboarding executes a Development Experience onboarding workflow for CLI commands.
func RunOnboarding(
	cmd *cobra.Command,
	assignment string,
	run func(context.Context, *Handle, string) (*development.DevelopmentOnboardingGuide, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, assignment)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatOnboarding(out), out)
}

// RunImpact executes a Development Experience impact workflow for CLI commands.
func RunImpact(
	cmd *cobra.Command,
	subject string,
	run func(context.Context, *Handle, string) (*development.DevelopmentImpactReport, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, subject)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatImpactReport(out), out)
}

// RunReuseCheck executes the Reuse Protocol workflow for CLI commands.
func RunReuseCheck(
	cmd *cobra.Command,
	intent string,
	run func(context.Context, *Handle, string) (*development.ReuseCheckResult, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, intent)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatReuseCheck(out), out)
}

// RunShipCheck executes the Ship Readiness workflow for CLI commands.
func RunShipCheck(
	cmd *cobra.Command,
	topic string,
	run func(context.Context, *Handle, string) (*development.ShipCheckResult, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, topic)
	if err != nil {
		return err
	}

	return WriteDEResult(cmd, development.FormatShipCheck(out), out)
}
