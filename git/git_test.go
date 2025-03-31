package git_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"go-kweb-lang/git"
)

type testCommandRunner struct {
	output string
	err    error

	// WorkingDir is the recorded workingDir
	WorkingDir string

	// Command is the recorded command with all arguments joined with space
	Command string
}

func (r *testCommandRunner) Exec(_ context.Context, workingDir string, cmd string, args ...string) (string, error) {
	r.WorkingDir = workingDir
	r.Command = cmd

	if len(args) > 0 {
		r.Command += " " + strings.Join(args, " ")
	}

	return r.output, r.err
}

func TestRepo_FindFileLastCommit(t *testing.T) {
	for _, tc := range []struct {
		name               string
		repoDir            string
		path               string
		runnerOutput       string
		runnerErr          error
		expectedWorkingDir string
		expectedCommand    string
		expectedErr        func(err error) bool
		expectedResult     git.CommitInfo
	}{
		{
			name:               "last commit does not exists",
			repoDir:            "/repo-dir",
			path:               "fake-file",
			runnerOutput:       "",
			runnerErr:          nil,
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git log -1 --format=%H %cd %s --date=iso-strict -- fake-file",
			expectedErr: func(err error) bool {
				return err == nil
			},
		},
		{
			name:               "last commit exists",
			repoDir:            "/repo-dir",
			path:               "content/pl/OWNERS",
			runnerOutput:       "d026267274d476357eee48df866fbba9e8875bb6 2024-05-18T02:04:55+09:00 update: OWNERS\n",
			runnerErr:          nil,
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git log -1 --format=%H %cd %s --date=iso-strict -- content/pl/OWNERS",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: git.CommitInfo{
				CommitID: "d026267274d476357eee48df866fbba9e8875bb6",
				DateTime: "2024-05-18T02:04:55+09:00",
				Comment:  "update: OWNERS",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testRunner := &testCommandRunner{
				output: tc.runnerOutput, err: tc.runnerErr,
			}

			repo := git.NewRepo(tc.repoDir, func(config *git.NewRepoConfig) {
				config.Runner = testRunner
			})

			commitInfo, err := repo.FindFileLastCommit(context.Background(), tc.path)

			if tc.expectedWorkingDir != testRunner.WorkingDir {
				t.Errorf("unexpected working dir\nactual   : %+v\nexptected: %+v",
					testRunner.WorkingDir, tc.expectedWorkingDir)
			}
			if tc.expectedCommand != testRunner.Command {
				t.Errorf("unexpected command\nactual   : %+v\nexptected: %+v",
					testRunner.Command, tc.expectedCommand)
			}
			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedResult != commitInfo {
				t.Errorf("unexpected result\nactual    : %+v\nexptected: %+v", &commitInfo, &tc.expectedResult)
			}
		})
	}
}

func TestRepo_FindFileCommitsAfter(t *testing.T) {
	for _, tc := range []struct {
		name               string
		repoDir            string
		path               string
		commitIDAfter      string
		runnerOutput       string
		runnerErr          error
		expectedWorkingDir string
		expectedCommand    string
		expectedErr        func(err error) bool
		expectedResult     []git.CommitInfo
	}{
		{
			name:               "no result",
			repoDir:            "/repo-dir",
			path:               "content/en/OWNERS",
			commitIDAfter:      "d026267274d476357eee48df866fbba9e8875bb6",
			runnerOutput:       "\n",
			runnerErr:          nil,
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git log --pretty=format:%H %cd %s --date=iso-strict d026267274d476357eee48df866fbba9e8875bb6.. -- content/en/OWNERS",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: nil,
		},
		{
			name:               "single line result",
			repoDir:            "/repo-dir",
			path:               "content/en/_index.html",
			commitIDAfter:      "f9120a9e5d2322cd3fc82db6417eb2fb77669a88",
			runnerOutput:       "e49c25cc17e83927498b1a7cbaa832e9100b5f36 2025-01-08T16:49:50+07:00 Redesign KubeCon links on the main page\n",
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git log --pretty=format:%H %cd %s --date=iso-strict f9120a9e5d2322cd3fc82db6417eb2fb77669a88.. -- content/en/_index.html",
			runnerErr:          nil,
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: []git.CommitInfo{
				{
					CommitID: "e49c25cc17e83927498b1a7cbaa832e9100b5f36",
					DateTime: "2025-01-08T16:49:50+07:00",
					Comment:  "Redesign KubeCon links on the main page",
				},
			},
		},
		{
			name:          "multiple lines result",
			repoDir:       "/repo-dir",
			path:          "content/en/docs/concepts/overview/_index.md",
			commitIDAfter: "f9120a9e5d2322cd3fc82db6417eb2fb77669a88",
			runnerOutput: "7e71096044ece8631e56439ea6ad3b6c456fb8a1 2024-09-11T15:17:51+03:00 Removed duplicated paragraph\n" +
				"f79eee0d92bda7b3f05931738b31de1f968efcc5 2024-08-26T06:58:06+08:00 Update _index.md\n",
			runnerErr:          nil,
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git log --pretty=format:%H %cd %s --date=iso-strict f9120a9e5d2322cd3fc82db6417eb2fb77669a88.. -- content/en/docs/concepts/overview/_index.md",
			expectedErr: func(err error) bool {
				return err == nil
			},
			expectedResult: []git.CommitInfo{
				{
					CommitID: "7e71096044ece8631e56439ea6ad3b6c456fb8a1",
					DateTime: "2024-09-11T15:17:51+03:00",
					Comment:  "Removed duplicated paragraph",
				},
				{
					CommitID: "f79eee0d92bda7b3f05931738b31de1f968efcc5",
					DateTime: "2024-08-26T06:58:06+08:00",
					Comment:  "Update _index.md",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testRunner := &testCommandRunner{
				output: tc.runnerOutput, err: tc.runnerErr,
			}

			repo := git.NewRepo(tc.repoDir, func(config *git.NewRepoConfig) {
				config.Runner = testRunner
			})

			commits, err := repo.FindFileCommitsAfter(context.Background(), tc.path, tc.commitIDAfter)

			if tc.expectedWorkingDir != testRunner.WorkingDir {
				t.Errorf("unexpected working dir\nactual   : %+v\nexptected: %+v",
					testRunner.WorkingDir, tc.expectedWorkingDir)
			}
			if tc.expectedCommand != testRunner.Command {
				t.Errorf("unexpected command\nactual   : %+v\nexptected: %+v",
					testRunner.Command, tc.expectedCommand)
			}
			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, tc.expectedResult) {
				t.Errorf("unexpected result\nactual   : %+v\nexptected: %+v", commits, tc.expectedResult)
			}
		})
	}
}

func TestLocalRepo_FindMergePoints(t *testing.T) {
	for _, tc := range []struct {
		name               string
		repoDir            string
		commitID           string
		runnerOutput       string
		runnerErr          error
		expectedWorkingDir string
		expectedCommand    string
		expectedErr        func(err error) bool
		expectedResult     []git.CommitInfo
	}{
		{
			name:               "no result",
			repoDir:            "/repo-dir",
			commitID:           "e49c25cc17e83927498b1a7cbaa832e9100b5f36",
			runnerOutput:       "\n",
			runnerErr:          nil,
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git --no-pager log --ancestry-path --merges --pretty=format:%H %cd %s --date=iso-strict e49c25cc17e83927498b1a7cbaa832e9100b5f36..main",
			expectedErr:        noError,
			expectedResult:     nil,
		},
		{
			name:               "single line result",
			repoDir:            "/repo-dir",
			commitID:           "e49c25cc17e83927498b1a7cbaa832e9100b5f36",
			runnerOutput:       "0e3a062280a55be00b533b258ee0e4c5e1f99f9d 2025-01-13T02:32:33-08:00 Merge pull request #49167 from shurup/upgrade-kubecon-section\n",
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git --no-pager log --ancestry-path --merges --pretty=format:%H %cd %s --date=iso-strict e49c25cc17e83927498b1a7cbaa832e9100b5f36..main",
			runnerErr:          nil,
			expectedErr:        noError,
			expectedResult: []git.CommitInfo{
				{
					CommitID: "0e3a062280a55be00b533b258ee0e4c5e1f99f9d",
					DateTime: "2025-01-13T02:32:33-08:00",
					Comment:  "Merge pull request #49167 from shurup/upgrade-kubecon-section",
				},
			},
		},
		{
			name:     "multiple lines result",
			repoDir:  "/repo-dir",
			commitID: "e49c25cc17e83927498b1a7cbaa832e9100b5f36",
			runnerOutput: "620d7f276c96789938869c67660d2f2aed42db49 2025-01-13T02:36:32-08:00 Merge pull request #48756 from sftim/20241118_localize_sidebar_tree_text\n" +
				"d4ecebf3699b126405953b47ccfea43caba72a0b 2025-01-13T02:34:32-08:00 Merge pull request #49171 from yuto-kimura-g/fix/49116\n" +
				"0e3a062280a55be00b533b258ee0e4c5e1f99f9d 2025-01-13T02:32:33-08:00 Merge pull request #49167 from shurup/upgrade-kubecon-section\n",
			runnerErr:          nil,
			expectedWorkingDir: "/repo-dir",
			expectedCommand:    "git --no-pager log --ancestry-path --merges --pretty=format:%H %cd %s --date=iso-strict e49c25cc17e83927498b1a7cbaa832e9100b5f36..main",
			expectedErr:        noError,
			expectedResult: []git.CommitInfo{
				{
					CommitID: "620d7f276c96789938869c67660d2f2aed42db49",
					DateTime: "2025-01-13T02:36:32-08:00",
					Comment:  "Merge pull request #48756 from sftim/20241118_localize_sidebar_tree_text",
				},
				{
					CommitID: "d4ecebf3699b126405953b47ccfea43caba72a0b",
					DateTime: "2025-01-13T02:34:32-08:00",
					Comment:  "Merge pull request #49171 from yuto-kimura-g/fix/49116",
				},
				{
					CommitID: "0e3a062280a55be00b533b258ee0e4c5e1f99f9d",
					DateTime: "2025-01-13T02:32:33-08:00",
					Comment:  "Merge pull request #49167 from shurup/upgrade-kubecon-section",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testRunner := &testCommandRunner{
				output: tc.runnerOutput, err: tc.runnerErr,
			}

			repo := git.NewRepo(tc.repoDir, func(config *git.NewRepoConfig) {
				config.Runner = testRunner
			})

			commits, err := repo.FindMergePoints(context.Background(), tc.commitID)

			if tc.expectedWorkingDir != testRunner.WorkingDir {
				t.Errorf("unexpected working dir\nactual   : %+v\nexptected: %+v",
					testRunner.WorkingDir, tc.expectedWorkingDir)
			}
			if tc.expectedCommand != testRunner.Command {
				t.Errorf("unexpected command\nactual   : %+v\nexptected: %+v",
					testRunner.Command, tc.expectedCommand)
			}
			if !tc.expectedErr(err) {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(commits, tc.expectedResult) {
				t.Errorf("unexpected result\nactual   : %+v\nexptected: %+v", commits, tc.expectedResult)
			}
		})
	}
}

func TestLocalRepo_MainBranchCommits(t *testing.T) {
	for _, tc := range []struct {
		name           string
		runnerOutput   string
		runnerErr      error
		expectedResult []git.CommitInfo
	}{
		{
			name:           "no result",
			runnerOutput:   "\n",
			runnerErr:      nil,
			expectedResult: nil,
		},
		{
			name: "multiple lines result",
			runnerOutput: "620d7f276c96789938869c67660d2f2aed42db49 2025-01-13T02:36:32-08:00 Merge pull request #48756 from sftim/20241118_localize_sidebar_tree_text\n" +
				"d4ecebf3699b126405953b47ccfea43caba72a0b 2025-01-13T02:34:32-08:00 Merge pull request #49171 from yuto-kimura-g/fix/49116\n" +
				"0e3a062280a55be00b533b258ee0e4c5e1f99f9d 2025-01-13T02:32:33-08:00 Merge pull request #49167 from shurup/upgrade-kubecon-section\n",
			runnerErr: nil,
			expectedResult: []git.CommitInfo{
				{
					CommitID: "620d7f276c96789938869c67660d2f2aed42db49",
					DateTime: "2025-01-13T02:36:32-08:00",
					Comment:  "Merge pull request #48756 from sftim/20241118_localize_sidebar_tree_text",
				},
				{
					CommitID: "d4ecebf3699b126405953b47ccfea43caba72a0b",
					DateTime: "2025-01-13T02:34:32-08:00",
					Comment:  "Merge pull request #49171 from yuto-kimura-g/fix/49116",
				},
				{
					CommitID: "0e3a062280a55be00b533b258ee0e4c5e1f99f9d",
					DateTime: "2025-01-13T02:32:33-08:00",
					Comment:  "Merge pull request #49167 from shurup/upgrade-kubecon-section",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testRunner := &testCommandRunner{
				output: tc.runnerOutput, err: tc.runnerErr,
			}

			repoDir := "/repo-dir"

			repo := git.NewRepo(repoDir, func(config *git.NewRepoConfig) {
				config.Runner = testRunner
			})

			commits, err := repo.MainBranchCommits(context.Background())

			expectedWorkingDir := repoDir

			if expectedWorkingDir != testRunner.WorkingDir {
				t.Errorf("unexpected working dir\nactual   : %+v\nexptected: %+v",
					testRunner.WorkingDir, expectedWorkingDir)
			}

			expectedCommand := "git --no-pager log main --pretty=format:%H %cd %s --date=iso-strict --first-parent"

			if expectedCommand != testRunner.Command {
				t.Errorf("unexpected command\nactual   : %+v\nexptected: %+v",
					testRunner.Command, expectedCommand)
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(commits, tc.expectedResult) {
				t.Errorf("unexpected result\nactual   : %+v\nexptected: %+v", commits, tc.expectedResult)
			}
		})
	}
}

func noError(err error) bool {
	return err == nil
}
