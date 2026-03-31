package browser

import (
	"reflect"
	"testing"
)

func TestBuildLaunchArgsAppendsDefaultVerificationURLs(t *testing.T) {
	t.Parallel()

	baseArgs := []string{"--disable-sync"}
	got := BuildLaunchArgs(append([]string{}, baseArgs...), &Profile{})
	want := []string{
		"--disable-sync",
		"https://ippure.com/",
		"https://iplark.com/",
		"https://ping0.cc/",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildLaunchArgs result mismatch:\n got=%v\nwant=%v", got, want)
	}
}

func TestBuildLaunchArgsSkipsDefaultVerificationURLsAfterInitialLaunch(t *testing.T) {
	t.Parallel()

	baseArgs := []string{"--disable-sync"}
	got := BuildLaunchArgs(append([]string{}, baseArgs...), &Profile{InitialVerificationDone: true})

	if !reflect.DeepEqual(got, baseArgs) {
		t.Fatalf("BuildLaunchArgs should not append verification URLs after initial launch: got=%v want=%v", got, baseArgs)
	}
}
