package obs

import (
	"errors"
	"strings"
	"testing"
)

func TestGetBuildStatus_Success(t *testing.T) {
	mockXML := `<resultlist>
		<result project="openSUSE:Factory" repository="standard" arch="x86_64" code="succeeded" state="published">
			<status package="hello" code="succeeded"/>
		</result>
		<result project="openSUSE:Factory" repository="standard" arch="i586" code="failed" state="published">
			<status package="hello" code="failed"/>
		</result>
	</resultlist>`

	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) ([]byte, error) {
		return []byte(mockXML), nil
	}

	client := NewClient()
	results, err := client.GetBuildStatus("openSUSE:Factory", "hello")

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 build results, got %d", len(results))
	}

	expectedStatus := "succeeded"
	if results[0].Status != expectedStatus {
		t.Errorf("Result[0] Status: got %q, want %q", results[0].Status, expectedStatus)
	}
}

func TestGetBuildStatus_ExecutionFailure(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	mockError := errors.New("command failed")
	execCommand = func(name string, args ...string) ([]byte, error) {
		return nil, mockError
	}

	client := NewClient()
	_, err := client.GetBuildStatus("badproject", "badpackage")

	if err == nil {
		t.Fatal("Expected an error due to command failure, but got nil")
	}

	if !strings.Contains(err.Error(), mockError.Error()) {
		t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
	}
}
