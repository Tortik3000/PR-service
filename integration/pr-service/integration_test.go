package pr_service_test

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"

	api "github.com/Tortik3000/PR-service/generated/api/pr-client"
)

var db *sql.DB

const (
	userTableName             = "users"
	prTableName               = "pull_request"
	teamTableName             = "team"
	assignedReviewerTableName = "assigned_reviewer"
)

func TestMain(m *testing.M) {
	godotenv.Load(".env.test")

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	source := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		dbName,
	)

	var err error
	db, err = sql.Open("postgres", source)

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	code := m.Run()
	db.Close()
	os.Exit(code)
}

func cleanUp(t *testing.T) {
	t.Helper()

	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", userTableName))
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", teamTableName))
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", prTableName))
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", assignedReviewerTableName))
	require.NoError(t, err)
}

//func TestPullRequestReassignConsistency(t *testing.T) {
//	ctx := context.Background()
//	executable := getPRServiceExecutable(t)
//	restPort := findFreePort(t)
//	client := newRESTClient(t, restPort)
//
//	cmd := setupPRService(t, executable, restPort)
//	t.Cleanup(func() {
//		stopPRService(t, cmd)
//		cleanUp(t)
//	})
//
//}

func TestTeam(t *testing.T) {
	executable := getPRServiceExecutable(t)
	restPort := findFreePort(t)
	metricsPort := findFreePort(t)

	cmd := setupPRService(t, executable, restPort, metricsPort)
	t.Cleanup(func() {
		stopPRService(t, cmd)
		cleanUp(t)
	})

	client := newRESTClient(t, restPort)

	t.Run("add team", func(t *testing.T) {
		t.Cleanup(func() {
			cleanUp(t)
		})

		ctx := context.Background()

		reqTeam := api.Team{
			TeamName: "team1",
			Members: []api.TeamMember{
				{
					IsActive: true,
					UserId:   "u1",
					Username: "name",
				},
				{
					IsActive: false,
					UserId:   "u2",
					Username: "name",
				},
			},
		}

		addTeamResp, err := client.PostTeamAddWithResponse(
			ctx,
			reqTeam,
		)

		require.NoError(t, err)
		respTeam := addTeamResp.JSON201.Team
		require.Equal(t, respTeam.TeamName, reqTeam.TeamName)
		require.Equal(t, respTeam.Members, reqTeam.Members)

		addTeamResp, err = client.PostTeamAddWithResponse(
			ctx,
			reqTeam,
		)
		require.NoError(t, err)
		require.Equal(t, addTeamResp.JSON400.Error.Code, api.TEAMEXISTS)

		invalidTeams := []api.Team{
			{
				TeamName: "",
				Members: []api.TeamMember{
					{
						IsActive: true,
						UserId:   "u1",
						Username: "name",
					},
				},
			},
			{
				Members: []api.TeamMember{
					{
						IsActive: true,
						UserId:   "u1",
						Username: "name",
					},
				},
			},
			{
				TeamName: "name1",
				Members: []api.TeamMember{
					{
						IsActive: true,
						UserId:   "u1",
					},
				},
			},
			{
				TeamName: "name2",
				Members: []api.TeamMember{
					{
						IsActive: true,
						UserId:   "u1",
						Username: "",
					},
				},
			},
			{
				TeamName: "name3",
				Members:  []api.TeamMember{},
			},
		}

		for _, invalidTeam := range invalidTeams {
			addTeamResp, err = client.PostTeamAddWithResponse(
				ctx,
				invalidTeam,
			)
			fmt.Println(addTeamResp.HTTPResponse)
			require.NoError(t, err)
			require.Equal(t, addTeamResp.HTTPResponse.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("add get", func(t *testing.T) {
		t.Cleanup(func() {
			cleanUp(t)
		})

		ctx := context.Background()

		reqTeam := api.Team{
			TeamName: "team2",
			Members: []api.TeamMember{
				{
					IsActive: true,
					UserId:   "u1",
					Username: "name",
				},
				{
					IsActive: true,
					UserId:   "u2",
					Username: "name",
				},
			},
		}

		_, err := client.PostTeamAddWithResponse(
			ctx,
			reqTeam,
		)
		require.NoError(t, err)

		getTeamResp, err := client.GetTeamGetWithResponse(
			ctx,
			&api.GetTeamGetParams{
				TeamName: reqTeam.TeamName,
			},
		)

		require.NoError(t, err)
		require.Equal(t, getTeamResp.JSON200.TeamName, reqTeam.TeamName)
		require.Equal(t, getTeamResp.JSON200.Members, reqTeam.Members)

		getTeamResp, err = client.GetTeamGetWithResponse(
			ctx,
			&api.GetTeamGetParams{
				TeamName: "not exist",
			},
		)
		require.NoError(t, err)
		require.Equal(t, getTeamResp.JSON404.Error.Code, api.NOTFOUND)

		invalidResp, err := client.GetTeamGetWithResponse(
			ctx,
			&api.GetTeamGetParams{
				TeamName: "",
			},
		)
		require.NoError(t, err)
		require.Equal(t, invalidResp.HTTPResponse.StatusCode, http.StatusBadRequest)
	})

}

func TestUser(t *testing.T) {
	executable := getPRServiceExecutable(t)
	restPort := findFreePort(t)
	metricsPort := findFreePort(t)

	cmd := setupPRService(t, executable, restPort, metricsPort)
	t.Cleanup(func() {
		stopPRService(t, cmd)
		cleanUp(t)
	})

	client := newRESTClient(t, restPort)

	t.Run("set user active", func(t *testing.T) {
		t.Cleanup(func() {
			cleanUp(t)
		})

		ctx := context.Background()

		reqUser := api.TeamMember{
			UserId:   "setIsActiveUser1",
			Username: "user1",
			IsActive: true,
		}

		reqTeam := api.Team{
			TeamName: "teamGetUserReview",
			Members:  []api.TeamMember{reqUser},
		}

		_, err := client.PostTeamAddWithResponse(ctx, reqTeam)
		require.NoError(t, err)

		resp, err := client.PostUsersSetIsActiveWithResponse(ctx, api.PostUsersSetIsActiveJSONRequestBody{
			UserId:   reqUser.UserId,
			IsActive: reqUser.IsActive,
		})
		require.NoError(t, err)
		require.Equal(t, resp.JSON200.User.UserId, reqUser.UserId)
		require.Equal(t, resp.JSON200.User.IsActive, reqUser.IsActive)

		invalidResp, err := client.PostUsersSetIsActiveWithResponse(ctx, api.PostUsersSetIsActiveJSONRequestBody{
			UserId:   "not_exist",
			IsActive: true,
		})
		require.NoError(t, err)
		require.Equal(t, invalidResp.JSON404.Error.Code, api.NOTFOUND)
	})

	t.Run("get user review", func(t *testing.T) {
		t.Cleanup(func() {
			cleanUp(t)
		})

		ctx := context.Background()

		user1 := api.TeamMember{
			IsActive: true,
			UserId:   "userRev1",
			Username: "name",
		}
		user2 := api.TeamMember{
			IsActive: true,
			UserId:   "userRev2",
			Username: "name",
		}

		shortPR := api.PullRequestShort{
			PullRequestId:   "userRev1",
			PullRequestName: "userRev1",
			AuthorId:        user1.UserId,
			Status:          api.PullRequestShortStatusOPEN,
		}
		reqTeam := api.Team{
			TeamName: "teamGetUserReview",
			Members:  []api.TeamMember{user1, user2},
		}

		_, err := client.PostTeamAddWithResponse(ctx, reqTeam)
		require.NoError(t, err)

		_, err = client.PostPullRequestCreateWithResponse(
			ctx,
			api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        shortPR.AuthorId,
				PullRequestId:   shortPR.PullRequestId,
				PullRequestName: shortPR.PullRequestName,
			},
		)
		require.NoError(t, err)

		getResp, err := client.GetUsersGetReviewWithResponse(ctx, &api.GetUsersGetReviewParams{
			UserId: user2.UserId,
		})
		require.NoError(t, err)
		require.Equal(t, getResp.JSON200.UserId, user2.UserId)
		require.Equal(t, getResp.JSON200.PullRequests, []api.PullRequestShort{shortPR})
	})
}

func TestPR(t *testing.T) {
	executable := getPRServiceExecutable(t)
	restPort := findFreePort(t)
	metricsPort := findFreePort(t)

	cmd := setupPRService(t, executable, restPort, metricsPort)
	t.Cleanup(func() {
		stopPRService(t, cmd)
		cleanUp(t)
	})

	client := newRESTClient(t, restPort)

	t.Run("merge pull request", func(t *testing.T) {
		t.Cleanup(func() {
			cleanUp(t)
		})

		ctx := context.Background()

		user1 := api.TeamMember{
			IsActive: true,
			UserId:   "prMerge1",
			Username: "name",
		}
		user2 := api.TeamMember{
			IsActive: true,
			UserId:   "prMerge2",
			Username: "name",
		}

		pr := api.PullRequest{
			PullRequestId:   "prMerge1",
			PullRequestName: "prMerge1",
			AuthorId:        user1.UserId,
			Status:          api.PullRequestStatusMERGED,
		}
		reqTeam := api.Team{
			TeamName: "prMerge1",
			Members:  []api.TeamMember{user1, user2},
		}

		_, err := client.PostTeamAddWithResponse(ctx, reqTeam)
		require.NoError(t, err)

		_, err = client.PostPullRequestCreateWithResponse(
			ctx,
			api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        pr.AuthorId,
				PullRequestId:   pr.PullRequestId,
				PullRequestName: pr.PullRequestName,
			},
		)
		require.NoError(t, err)

		response, err := client.PostPullRequestMergeWithResponse(
			ctx,
			api.PostPullRequestMergeJSONRequestBody{
				PullRequestId: pr.PullRequestId,
			})

		require.NoError(t, err)
		require.Equal(t, response.JSON200.Pr.Status, pr.Status)

		response, err = client.PostPullRequestMergeWithResponse(
			ctx,
			api.PostPullRequestMergeJSONRequestBody{
				PullRequestId: "not exist",
			})

		require.NoError(t, err)
		require.Equal(t, response.JSON404.Error.Code, api.NOTFOUND)
	})

	t.Run("reassign pull request", func(t *testing.T) {
		t.Cleanup(func() {
			cleanUp(t)
		})

		ctx := context.Background()

		user1 := api.TeamMember{
			IsActive: true,
			UserId:   "reassignPR1",
			Username: "name",
		}
		user2 := api.TeamMember{
			IsActive: true,
			UserId:   "reassignPR2",
			Username: "name",
		}
		user3 := api.TeamMember{
			IsActive: true,
			UserId:   "reassignPR3",
			Username: "name",
		}
		user4 := api.TeamMember{
			IsActive: false,
			UserId:   "reassignPR4",
			Username: "name",
		}

		pr := api.PullRequest{
			PullRequestId:   "reassignPR1",
			PullRequestName: "reassignPR1",
			AuthorId:        user1.UserId,
			Status:          api.PullRequestStatusMERGED,
		}
		reqTeam := api.Team{
			TeamName: "reassignPR1",
			Members:  []api.TeamMember{user1, user2, user3, user4},
		}

		_, err := client.PostTeamAddWithResponse(ctx, reqTeam)
		require.NoError(t, err)

		response, err := client.PostPullRequestCreateWithResponse(
			ctx,
			api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        pr.AuthorId,
				PullRequestId:   pr.PullRequestId,
				PullRequestName: pr.PullRequestName,
			},
		)
		require.NoError(t, err)
		reviewers := []string{user2.UserId, user3.UserId}
		slices.Sort(reviewers)
		slices.Sort(response.JSON201.Pr.AssignedReviewers)
		require.Equal(t, response.JSON201.Pr.AssignedReviewers, reviewers)

		reassignResp, err := client.PostPullRequestReassignWithResponse(
			ctx,
			api.PostPullRequestReassignJSONRequestBody{
				PullRequestId: pr.PullRequestId,
				OldUserId:     user2.UserId,
			})

		fmt.Println(reassignResp.JSON409.Error.Code)
		require.NoError(t, err)
		require.Equal(t, reassignResp.JSON409.Error.Code, api.NOCANDIDATE)

		_, err = client.PostUsersSetIsActiveWithResponse(
			ctx,
			api.PostUsersSetIsActiveJSONRequestBody{
				UserId:   user4.UserId,
				IsActive: true,
			})
		require.NoError(t, err)

		reassignResp, err = client.PostPullRequestReassignWithResponse(
			ctx,
			api.PostPullRequestReassignJSONRequestBody{
				PullRequestId: pr.PullRequestId,
				OldUserId:     user2.UserId,
			})

		require.NoError(t, err)
		require.Equal(t, reassignResp.JSON200.ReplacedBy, user4.UserId)

		reviewers = []string{user4.UserId, user3.UserId}
		slices.Sort(reviewers)
		slices.Sort(reassignResp.JSON200.Pr.AssignedReviewers)

		require.Equal(t, reviewers, reassignResp.JSON200.Pr.AssignedReviewers, reviewers)

		_, err = client.PostPullRequestMergeWithResponse(
			ctx,
			api.PostPullRequestMergeJSONRequestBody{
				PullRequestId: pr.PullRequestId,
			})

		require.NoError(t, err)

		reassignResp, err = client.PostPullRequestReassignWithResponse(
			ctx,
			api.PostPullRequestReassignJSONRequestBody{
				PullRequestId: pr.PullRequestId,
				OldUserId:     user3.UserId,
			})

		require.NoError(t, err)
		require.Equal(t, reassignResp.JSON409.Error.Code, api.PRMERGED)
	})
}

var requiredEnv = []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD"}

func setupPRService(
	t *testing.T,
	executable string,
	restPort string,
	metricsPort string,
) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for _, p := range requiredEnv {
		cur := os.Getenv(p)
		require.NotEmpty(t, cur, "you need to pass env variable to tests: "+p)

		cmd.Env = append(cmd.Env, p+"="+cur)
	}

	cmd.Env = append(cmd.Env, "REST_PORT="+restPort)
	cmd.Env = append(cmd.Env, "METRICS_PORT="+metricsPort)

	require.NoError(t, cmd.Start())
	restClient := newRESTClient(t, restPort)

	// rest health check
	for i := range 50 {

		_, err := restClient.PostTeamAddWithResponse(context.Background(), api.Team{
			TeamName: "healthcheck",
			Members: []api.TeamMember{
				{
					IsActive: true,
					UserId:   "u1",
					Username: "name",
				},
				{
					IsActive: false,
					UserId:   "u2",
					Username: "name",
				},
			},
		})

		_, ok := status.FromError(err)

		if ok {
			break
		}

		if i == 49 {
			require.NoError(t, err)
			log.Println("rest health check error")
			t.Fail()
		}

		time.Sleep(time.Millisecond * 100)
	}

	return cmd
}

func stopPRService(t *testing.T, cmd *exec.Cmd) {
	t.Helper()

	for i := 0; i < 5; i++ {
		require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))
	}

	require.NoError(t, cmd.Wait())
	require.Equal(t, 0, cmd.ProcessState.ExitCode())
}

func getPRServiceExecutable(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	binaryPath, err := resolveFilePath(filepath.Dir(filepath.Dir(wd)), "pr-service")
	require.NoError(t, err, "you need to compile your pr-service, run make build")

	return binaryPath
}

func newRESTClient(t *testing.T, prot string) api.ClientWithResponsesInterface {
	t.Helper()

	addr := "http://localhost:" + prot
	client, err := api.NewClientWithResponses(addr)
	require.NoError(t, err)

	return client
}

func findFreePort(t *testing.T) string {
	t.Helper()

	for {
		port := rand.N(16383) + 49152
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)

		if err == nil {
			require.NoError(t, ln.Close())
			return strconv.Itoa(port)
		}
	}
}

func resolveFilePath(root string, filename string) (string, error) {
	cleanedRoot := filepath.Clean(root)
	nameWithoutExt := strings.TrimRight(root, filepath.Ext(filename))

	var result string

	err := filepath.WalkDir(cleanedRoot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		name := d.Name()

		if name == filename || name == nameWithoutExt {
			result = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("walk fail tree fail, error: %w", err)
	}

	if result == "" {
		return "", fmt.Errorf("file %s not found in root %s", filename, root)
	}

	return result, nil
}
