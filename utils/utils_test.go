package utils

import (
	"fmt"
	"testing"

	exutil "gitlab-release.hortorgames.com/platform/statis.utils/utils"
)

func TestHashCode(t *testing.T) {
	extraID := "odBxvt78QS0_4JsnEwdZ3mpBoK9s"
	fmt.Printf("id: %d\n", exutil.HashCode(extraID)%30)
}
