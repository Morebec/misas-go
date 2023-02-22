package spectool

import (
	"github.com/morebec/specter"
	"testing"
)

func TestSpecificationTool(t *testing.T) {
	tool := New(specter.FullMode)
	if err := tool.Run([]string{"./test_data"}); err != nil {
		panic(err)
	}
}
