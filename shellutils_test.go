package easyssh

import (
	"testing"
	"fmt"
)

func TestRtLocal(t *testing.T) {
	RtLocal("ls -lh /tmp", func(line string, lineType int8) {
		fmt.Println(line)
	})
}
