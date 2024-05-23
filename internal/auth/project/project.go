package project

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/llm-operator/cli/internal/auth/org"
	ihttp "github.com/llm-operator/cli/internal/http"
	"github.com/llm-operator/cli/internal/runtime"
	"github.com/llm-operator/cli/internal/ui"
	uv1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/pkg/role"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

const (
	pathPattern = "/organizations/%s/projects"
)

// Cmd is the root command for projectanizations.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "projects",
		Short:              "projects commands",
		Aliases:            []string{"project"},
		Args:               cobra.NoArgs,
		DisableFlagParsing: true,
	}

	cmd.AddCommand(createCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(deleteCmd())
	cmd.AddCommand(addMemberCmd())
	cmd.AddCommand(listMembersCmd())
	cmd.AddCommand(removeMemberCmd())

	return cmd
}

func createCmd() *cobra.Command {
	var title, orgTitle, namespace string
	cmd := &cobra.Command{
		Use:  "create",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return create(cmd.Context(), title, orgTitle, namespace)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title of the project")
	cmd.Flags().StringVarP(&orgTitle, "organization-title", "o", "", "Organization title of the project. The organization in the current context is used if not specified.")
	cmd.Flags().StringVarP(&namespace, "kubernetes-namespace", "n", "", "Kubernetes namesapce of the project")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("kubernetes-namespace")
	return cmd
}

func listCmd() *cobra.Command {
	var orgTitle string
	cmd := &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return list(cmd.Context(), orgTitle)
		},
	}
	cmd.Flags().StringVarP(&orgTitle, "organization-title", "o", "", "Organization title of the project. The organization in the current context is used if not specified.")
	return cmd
}

func deleteCmd() *cobra.Command {
	var title, orgTitle string
	cmd := &cobra.Command{
		Use:  "delete",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return delete(cmd.Context(), title, orgTitle)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title of the project")
	cmd.Flags().StringVarP(&orgTitle, "organization-title", "o", "", "Organization title of the project. The organization in the current context is used if not specified.")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func addMemberCmd() *cobra.Command {
	var title, orgTitle, email, roleStr string
	cmd := &cobra.Command{
		Use:  "add-member",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, ok := role.ProjectRoleToProtoEnum(roleStr)
			if !ok {
				return fmt.Errorf("invalid role %q", roleStr)
			}
			return addMember(cmd.Context(), title, orgTitle, email, r)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title of the project")
	cmd.Flags().StringVarP(&orgTitle, "organization-title", "o", "", "Organization title of the project. The organization in the current context is used if not specified.")
	cmd.Flags().StringVar(&email, "email", "", "Email of the user")
	cmd.Flags().StringVar(&roleStr, "role", "", "Role of the user (owner or reader)")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func listMembersCmd() *cobra.Command {
	var title, orgTitle string
	cmd := &cobra.Command{
		Use:  "list-members",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listMembers(cmd.Context(), title, orgTitle)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title of the project")
	cmd.Flags().StringVarP(&orgTitle, "organization-title", "o", "", "Organization title of the project. The organization in the current context is used if not specified.")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}

func removeMemberCmd() *cobra.Command {
	var title, orgTitle, email string
	cmd := &cobra.Command{
		Use:  "remove-member",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeMember(cmd.Context(), title, orgTitle, email)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title of the project")
	cmd.Flags().StringVarP(&orgTitle, "organization-title", "o", "", "Organization title of the project. The organization in the current context is used if not specified.")
	cmd.Flags().StringVar(&email, "email", "", "Email of the user")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func create(ctx context.Context, title, orgTitle, namespace string) error {
	env, err := runtime.NewEnv(ctx)
	if err != nil {
		return err
	}

	path, orgID, err := buildPath(env, orgTitle)
	if err != nil {
		return err
	}
	req := uv1.CreateProjectRequest{
		Title:               title,
		OrganizationId:      orgID,
		KubernetesNamespace: namespace,
	}
	var resp uv1.Project
	if err := ihttp.NewClient(env).Send(http.MethodPost, path, &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Created a new project. (ID: %s)\n", resp.Id)
	return nil
}

func list(ctx context.Context, orgTitle string) error {
	env, err := runtime.NewEnv(ctx)
	if err != nil {
		return err
	}

	projects, err := ListProjects(env, orgTitle)
	if err != nil {
		return err
	}

	tbl := table.New("Title", "Kubernetes namespace", "Created At")
	ui.FormatTable(tbl)

	for _, o := range projects {
		tbl.AddRow(o.Title, o.KubernetesNamespace, time.Unix(o.CreatedAt, 0).Format(time.RFC3339))
	}

	tbl.Print()

	return nil
}

func delete(ctx context.Context, title, orgTitle string) error {
	env, err := runtime.NewEnv(ctx)
	if err != nil {
		return err
	}

	project, found, err := FindProjectByTitle(env, title, orgTitle)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("project %q not found", title)
	}

	path, orgID, err := buildPath(env, orgTitle)
	if err != nil {
		return err
	}
	req := uv1.DeleteProjectRequest{
		Id:             project.Id,
		OrganizationId: orgID,
	}
	var resp uv1.DeleteProjectResponse
	if err := ihttp.NewClient(env).Send(http.MethodDelete, fmt.Sprintf("%s/%s", path, project.Id), &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Deleted the project (ID: %q).\n", project.Id)

	return nil
}

func addMember(ctx context.Context, title, orgTitle, userID string, role uv1.ProjectRole) error {
	env, err := runtime.NewEnv(ctx)
	if err != nil {
		return err
	}

	project, found, err := FindProjectByTitle(env, title, orgTitle)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("project %q not found", title)
	}

	path, orgID, err := buildPath(env, orgTitle)
	if err != nil {
		return err
	}
	req := uv1.CreateProjectUserRequest{
		ProjectId:      project.Id,
		OrganizationId: orgID,
		UserId:         userID,
		Role:           role,
	}
	var resp uv1.ProjectUser
	p := fmt.Sprintf("%s/%s/users", path, project.Id)
	if err := ihttp.NewClient(env).Send(http.MethodPost, p, &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Added the user %q to the project %q with role %q.\n", userID, title, role.String())

	return nil
}

func listMembers(ctx context.Context, title, orgTitle string) error {
	env, err := runtime.NewEnv(ctx)
	if err != nil {
		return err
	}

	project, found, err := FindProjectByTitle(env, title, orgTitle)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("project %q not found", title)
	}

	path, orgID, err := buildPath(env, orgTitle)
	if err != nil {
		return err
	}
	req := uv1.ListProjectUsersRequest{
		ProjectId:      project.Id,
		OrganizationId: orgID,
	}
	var resp uv1.ListProjectUsersResponse
	p := fmt.Sprintf("%s/%s/users", path, project.Id)
	if err := ihttp.NewClient(env).Send(http.MethodGet, p, &req, &resp); err != nil {
		return err
	}

	tbl := table.New("User ID", "Role")
	ui.FormatTable(tbl)
	for _, u := range resp.Users {
		tbl.AddRow(u.UserId, u.Role.String())
	}

	tbl.Print()

	return nil
}

func removeMember(ctx context.Context, title, orgTitle, userID string) error {
	env, err := runtime.NewEnv(ctx)
	if err != nil {
		return err
	}

	project, found, err := FindProjectByTitle(env, title, orgTitle)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("project %q not found", title)
	}

	path, orgID, err := buildPath(env, orgTitle)
	if err != nil {
		return err
	}
	req := uv1.DeleteProjectUserRequest{
		ProjectId:      project.Id,
		OrganizationId: orgID,
		UserId:         userID,
	}
	var resp uv1.ProjectUser
	p := fmt.Sprintf("%s/%s/users/%s", path, project.Id, userID)
	if err := ihttp.NewClient(env).Send(http.MethodDelete, p, &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Removed the user %q from the project %q.\n", userID, title)

	return nil
}

// FindProjectByTitle finds a project by title.
func FindProjectByTitle(env *runtime.Env, title, orgTitle string) (*uv1.Project, bool, error) {
	projects, err := ListProjects(env, orgTitle)
	if err != nil {
		return nil, false, err
	}
	for _, p := range projects {
		if p.Title == title {
			return p, true, nil
		}
	}
	return nil, false, nil
}

// ListProjects lists projects.
func ListProjects(env *runtime.Env, orgTitle string) ([]*uv1.Project, error) {
	path, orgID, err := buildPath(env, orgTitle)
	if err != nil {
		return nil, err
	}

	req := uv1.ListProjectsRequest{
		OrganizationId: orgID,
	}
	var resp uv1.ListProjectsResponse
	if err := ihttp.NewClient(env).Send(http.MethodGet, path, &req, &resp); err != nil {
		return nil, err
	}
	return resp.Projects, nil
}

func buildPath(env *runtime.Env, orgTitle string) (string, string, error) {
	orgID, err := getOrgID(env, orgTitle)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf(pathPattern, orgID), orgID, nil
}

func getOrgID(env *runtime.Env, orgTitle string) (string, error) {
	if orgTitle == "" {
		oid := env.Config.Context.OrganizationID
		if oid == "" {
			return "", fmt.Errorf("--organization-title flag must be specified or the organization must be specified in the configuration file")
		}
		return oid, nil
	}

	org, found, err := org.FindOrgByTitle(env, orgTitle)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("organization %q not found", orgTitle)
	}
	return org.Id, nil
}
