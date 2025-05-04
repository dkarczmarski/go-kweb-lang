package git_test

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"go-kweb-lang/git"
)

func TestLocalRepo_Create_Integration(t *testing.T) {
	for _, tc := range []struct {
		name        string
		repoURL     func(repoDirPath string) string
		expectedErr func(err error) bool
	}{
		{
			name: "creates repo from correct url",
			repoURL: func(repoDirPath string) string {
				return "file://" + repoDirPath
			},
			expectedErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "creates repo from incorrect url",
			repoURL: func(repoDirPath string) string {
				return "_bad_url_" + repoDirPath
			},
			expectedErr: func(err error) bool {
				return err != nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			srcRepoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)
			repoURL := tc.repoURL(srcRepoPath)

			repoPath := filepath.Join(tmpDir, "test-repo")
			if err := os.Mkdir(repoPath, 0o755); err != nil {
				t.Fatalf("failed to create directory %s: %v", repoPath, err)
			}

			gitRepo := git.NewRepo(repoPath)

			err := gitRepo.Create(ctx, repoURL)
			if !tc.expectedErr(err) {
				t.Errorf("unexpected errror when creating git repo at %s from %s: %v", repoPath, repoURL, err)
			}
		})
	}
}

func TestLocalRepo_Checkout_Integration(t *testing.T) {
	for _, tc := range []struct {
		name     string
		commitID string
		checkErr func(err error) bool
	}{
		{
			name:     "when commitID is correct",
			commitID: "2a2c911b4a8e0e681dafa7b236446a9e42f47533",
			checkErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name:     "when commitID does not exist",
			commitID: "AAAAA",
			checkErr: func(err error) bool {
				return err != nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			commitID := tc.commitID
			err := gitRepo.Checkout(ctx, commitID)

			if !tc.checkErr(err) {
				t.Errorf("unexpected error when doing checkout commit %s: %v", commitID, err)
			}
		})
	}
}

func TestLocalRepo_MainBranchCommits_Integration(t *testing.T) {
	for _, tc := range []struct {
		name            string
		expectedErr     func(err error) bool
		expectedCommits []git.CommitInfo
	}{
		{
			name: "checking initialized repo returns commits",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedCommits: []git.CommitInfo{
				{
					CommitID: "91decf29e674f6faf478f9641b183725569382d0",
					DateTime: "2020-01-15T00:00:00+00:00",
					Comment:  "Merge branch 'branch3'",
				},
				{
					CommitID: "5de1d14f917738d148657f1c8c198bfd8f2c5a27",
					DateTime: "2020-01-14T00:00:00+00:00",
					Comment:  "Merge branch 'branch1'",
				},
				{
					CommitID: "97056283dccdf83d7a2994e58684048d697d9ba0",
					DateTime: "2020-01-13T00:00:00+00:00",
					Comment:  "commit (main) file3.txt",
				},
				{
					CommitID: "2a2c911b4a8e0e681dafa7b236446a9e42f47533",
					DateTime: "2020-01-05T00:00:00+00:00",
					Comment:  "commit (main) file1.txt 2",
				},
				{
					CommitID: "40eda29f9d285779f07b474a02920b6a379d8af0",
					DateTime: "2020-01-04T00:00:00+00:00",
					Comment:  "commit (main) file2.txt",
				},
				{
					CommitID: "a706ad7f9b265094aaad26b512ece0787f09c652",
					DateTime: "2020-01-03T00:00:00+00:00",
					Comment:  "commit (main) file1.txt",
				},
				{
					CommitID: "754fcc8e8ba56078eb3a4ebc5cfbbfe3201e4ea4",
					DateTime: "2020-01-02T00:00:00+00:00",
					Comment:  "init commit",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			commits, err := gitRepo.ListMainBranchCommits(ctx)

			if !tc.expectedErr(err) {
				t.Errorf("unexptected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedCommits, commits) {
				t.Errorf("unexpected result. expected: %+v\nactual  : %+v", tc.expectedCommits, commits)
			}
		})
	}
}

func TestLocalRepo_FileExists_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		fileName       string
		expectedErr    func(err error) bool
		expectedExists bool
	}{
		{
			name:     "when file exists",
			fileName: "file1.txt",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedExists: true,
		},
		{
			name:     "when file does not exist",
			fileName: "fake-file.txt",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedExists: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			exists, err := gitRepo.FileExists(tc.fileName)

			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if tc.expectedExists != exists {
				t.Errorf("unexpected result: %v", exists)
			}
		})
	}
}

func TestLocalRepo_ListFiles_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		dirPath        string
		expectedErr    func(err error) bool
		expectedResult []string
	}{
		{
			name:    "list files at initialized repo",
			dirPath: ".",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: []string{
				"file1.txt",
				"file2.txt",
				"file3.txt",
				"file4.txt",
				"file5.txt",
				"file6.txt",
				"file7.txt",
				"file8.txt",
				"init.txt",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			files, err := gitRepo.ListFiles(tc.dirPath)

			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedResult, files) {
				t.Errorf("unexpected result: %v", files)
			}
		})
	}
}

func TestLocalRepo_FindFileLastCommit_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		filePath       string
		expectedErr    func(err error) bool
		expectedResult git.CommitInfo
	}{
		{
			name:     "when file exists",
			filePath: "file1.txt",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: git.CommitInfo{
				CommitID: "2a2c911b4a8e0e681dafa7b236446a9e42f47533",
				DateTime: "2020-01-05T00:00:00+00:00",
				Comment:  "commit (main) file1.txt 2",
			},
		},
		{
			// todo: probably it should be some "not-found" error
			name:     "when file does not exist",
			filePath: "fake-file",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: git.CommitInfo{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			commit, err := gitRepo.FindFileLastCommit(ctx, tc.filePath)

			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedResult, commit) {
				t.Errorf("unexpected result: %+v", commit)
			}
		})
	}
}

func TestLocalRepo_FindFileCommitsAfter_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		filePath       string
		commitIDFrom   string
		expectedResult []git.CommitInfo
	}{
		{
			name:         "when the file commits are found",
			filePath:     "file4.txt",
			commitIDFrom: "2a2c911b4a8e0e681dafa7b236446a9e42f47533",
			expectedResult: []git.CommitInfo{
				{
					CommitID: "8dd9d9f078564d285e1945ef75d17d87eef55c33",
					DateTime: "2020-01-09T00:00:00+00:00",
					Comment:  "commit (branch1) file4.txt 2",
				},
				{
					CommitID: "5a941744bc6dbdd39032cf076101f3f530afb295",
					DateTime: "2020-01-06T00:00:00+00:00",
					Comment:  "commit (branch1) file4.txt",
				},
			},
		},
		{
			name:           "when the file does not exist",
			filePath:       "fake-file",
			commitIDFrom:   "2a2c911b4a8e0e681dafa7b236446a9e42f47533",
			expectedResult: nil,
		},
		{
			name:           "when the file commits are not found",
			filePath:       "file4.txt",
			commitIDFrom:   "bd5b85e507387cb8ee4036ff1e672f1e64cdd618",
			expectedResult: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			result, err := gitRepo.FindFileCommitsAfter(ctx, tc.filePath, tc.commitIDFrom)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedResult, result) {
				t.Errorf("unexpected result: %+v", result)
			}
		})
	}
}

func TestLocalRepo_FindMergePoints_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		commitID       string
		expectedErr    func(err error) bool
		expectedResult []git.CommitInfo
	}{
		{
			name:     "when commitID exists",
			commitID: "fa952b36a05bc120001024566cec6b16914efe03",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: []git.CommitInfo{
				{
					CommitID: "91decf29e674f6faf478f9641b183725569382d0",
					DateTime: "2020-01-15T00:00:00+00:00",
					Comment:  "Merge branch 'branch3'",
				},
				{
					CommitID: "5de1d14f917738d148657f1c8c198bfd8f2c5a27",
					DateTime: "2020-01-14T00:00:00+00:00",
					Comment:  "Merge branch 'branch1'",
				},
				{
					CommitID: "01da95266faa3d418fa2c775ada3d9423ea1a598",
					DateTime: "2020-01-11T00:00:00+00:00",
					Comment:  "Merge branch 'branch2' into branch1",
				},
			},
		},
		{
			name:     "when commitID does not exist",
			commitID: "AAAA",
			expectedErr: func(err error) bool {
				return err != nil
			},
			expectedResult: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			result, err := gitRepo.ListMergePoints(ctx, tc.commitID)

			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedResult, result) {
				t.Errorf("unexpected result: %+v", result)
			}
		})
	}
}

func TestLocalRepo_Pull_Integration_Scenario(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	srcRepoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)
	repoURL := "file://" + srcRepoPath

	repoPath := filepath.Join(tmpDir, "test-repo")
	if err := os.Mkdir(repoPath, 0o755); err != nil {
		t.Fatalf("failed while initializing test: %v", err)
	}

	gitRepo := git.NewRepo(repoPath)

	if err := gitRepo.Create(ctx, repoURL); err != nil {
		t.Fatalf("failed while initializing test: %v", err)
	}

	_ = runInitRepoScript(t, tmpDir, "initrepo2.sh", initRepoScript2Content)

	t.Run("before performing fetch the file should not exist", func(t *testing.T) {
		exist, err := gitRepo.FileExists("file10.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exist {
			t.Fatal("file should not exist")
		}
	})

	t.Run("before performing fetch the list of fresh commits should be empty", func(t *testing.T) {
		freshCommits, err := gitRepo.ListFreshCommits(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(freshCommits) > 0 {
			t.Fatal("fresh commit list should be empty")
		}
	})

	t.Run("perform fetch", func(t *testing.T) {
		if err := gitRepo.Fetch(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("after performing fetch the list of fresh commits should not be empty", func(t *testing.T) {
		freshCommits, err := gitRepo.ListFreshCommits(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(freshCommits) == 0 {
			t.Fatal("fresh commit list should not be empty")
		}

		expectedCommits := []git.CommitInfo{
			{
				CommitID: "f10292d763daa46218f12aec9d02460fbbb540d9",
				DateTime: "2020-10-03T00:00:00+00:00",
				Comment:  "commit (main) file11.txt",
			},
			{
				CommitID: "c4f3c787c1d4af42433605de3accc0d3b3b60d2d",
				DateTime: "2020-10-02T00:00:00+00:00",
				Comment:  "commit (main) file10.txt",
			},
		}

		if !reflect.DeepEqual(expectedCommits, freshCommits) {
			t.Errorf("unexpected result: %+v", freshCommits)
		}
	})

	t.Run("before performing pull, the file should exist", func(t *testing.T) {
		exist, err := gitRepo.FileExists("file10.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exist {
			t.Fatal("file should not exist")
		}
	})

	t.Run("perform pull", func(t *testing.T) {
		if err := gitRepo.Pull(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("after performing pull the list of fresh commits should be empty", func(t *testing.T) {
		freshCommits, err := gitRepo.ListFreshCommits(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(freshCommits) > 0 {
			t.Fatal("fresh commit list should be empty")
		}
	})

	t.Run("after performing pull, the file should exist", func(t *testing.T) {
		exist, err := gitRepo.FileExists("file10.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exist {
			t.Fatal("file should exist")
		}
	})
}

func TestLocalRepo_FilesInCommit_Integration(t *testing.T) {
	for _, tc := range []struct {
		name           string
		commitID       string
		expectedErr    func(err error) bool
		expectedResult []string
	}{
		{
			name:     "when commit has one file",
			commitID: "40eda29f9d285779f07b474a02920b6a379d8af0",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: []string{
				"file2.txt",
			},
		},
		{
			name:     "when commit has two files",
			commitID: "fc3a8fa89964ba1bc2d19f2caa655f2037551ed7",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: []string{
				"file2.txt",
				"file7.txt",
			},
		},
		{
			name:     "when commit does not exist",
			commitID: "AAAA",
			expectedErr: func(err error) bool {
				return err != nil
			},
			expectedResult: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tmpDir := t.TempDir()

			repoPath := runInitRepoScript(t, tmpDir, "initrepo.sh", initRepoScriptContent)

			gitRepo := git.NewRepo(repoPath)

			result, err := gitRepo.ListFilesInCommit(ctx, tc.commitID)

			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedResult, result) {
				t.Errorf("unexpected result: %+v", result)
			}
		})
	}
}

//go:embed testdata/initrepo.sh
var initRepoScriptContent []byte

//go:embed testdata/initrepo2.sh
var initRepoScript2Content []byte

var initRepoScriptDebugOutput bool = true

func runInitRepoScript(t *testing.T, tmpDir string, scriptName string, scriptContent []byte) string {
	scriptPath := filepath.Join(tmpDir, scriptName)

	if err := os.WriteFile(scriptPath, scriptContent, 0o755); err != nil {
		t.Fatalf("failed to write script file: %v", err)
	}

	if err := os.Chmod(scriptPath, 0o755); err != nil {
		t.Fatalf("failed to chmod script file: %v", err)
	}

	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = tmpDir
	if initRepoScriptDebugOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		t.Fatalf("script execution failed: %v", err)
	}

	return filepath.Join(tmpDir, "repo")
}
