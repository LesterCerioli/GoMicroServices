package ctx

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/3Rivers/go-zero/core/stringx"
	"github.com/3Rivers/go-zero/tools/goctl/rpc/execx"
	"github.com/3Rivers/go-zero/tools/goctl/util"
	"github.com/stretchr/testify/assert"
)

func TestProjectFromGoMod(t *testing.T) {
	dft := build.Default
	gp := dft.GOPATH
	if len(gp) == 0 {
		return
	}
	projectName := stringx.Rand()
	dir := filepath.Join(gp, "src", projectName)
	err := util.MkdirIfNotExist(dir)
	if err != nil {
		return
	}

	_, err = execx.Run("go mod init "+projectName, dir)
	assert.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	ctx, err := projectFromGoMod(dir)
	assert.Nil(t, err)
	assert.Equal(t, projectName, ctx.Path)
	assert.Equal(t, dir, ctx.Dir)
}
