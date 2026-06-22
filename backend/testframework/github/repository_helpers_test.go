package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchTestCaseFileByPath(t *testing.T) {
	tests := []struct {
		name      string
		className string
		fullName  string
		filePath  string
		want      bool
	}{
		{
			name:      "class name match with slashes",
			className: "com.example.LoginTest",
			filePath:  "src/test/java/com/example/LoginTest.java",
			want:      true,
		},
		{
			name:     "full name match",
			fullName: "com.example.LoginTest",
			filePath: "src/test/java/com/example/LoginTest.java",
			want:     true,
		},
		{
			name:      "backslash path normalized",
			className: "com.example.AuthTest",
			filePath:  `src\test\java\com\example\AuthTest.java`,
			want:      true,
		},
		{
			name:      "no match",
			className: "com.other.Test",
			filePath:  "src/main/java/App.java",
			want:      false,
		},
		{
			name:     "empty identifiers",
			filePath: "src/App.java",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := github.ExportMatchTestCaseFileByPath(tt.className, tt.fullName, tt.filePath)
			assert.Equal(t, tt.want, got)
		})
	}
}
