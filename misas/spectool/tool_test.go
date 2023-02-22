package spectool

import (
	"testing"
)

func TestSpecificationTool(t *testing.T) {
	tool := SpecificationTool()
	if err := tool.Run([]string{"./test_data"}); err != nil {
		panic(err)
	}
}
