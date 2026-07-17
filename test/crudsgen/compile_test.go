//go:build crudsgen

package crudsgen

import (
	"testing"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/test/testmodels"
)

func TestUserRowInterface(t *testing.T) {
	var _ cruds.TypedTable[testmodels.UserRow] = (*testmodels.TableUserRow)(nil)
}

func TestProductRowInterface(t *testing.T) {
	var _ cruds.TypedTable[testmodels.ProductRow] = (*testmodels.TableProductRow)(nil)
}

func TestFuncRowInterface(t *testing.T) {
	var _ cruds.TypedTable[testmodels.FuncRow] = (*testmodels.TableFuncRow)(nil)
}

func TestIdNameAgeRowFilledInterface(t *testing.T) {
	var _ cruds.TypedTable[testmodels.IdNameAgeRowFilled] = (*testmodels.TableIdNameAgeRowFilled)(nil)
}

func TestJoinSampleInterface(t *testing.T) {
	var _ cruds.TypedJoiner[testmodels.JoinSample] = (*testmodels.JoinerJoinSample)(nil)
}

func TestJoinSummaryInterface(t *testing.T) {
	var _ cruds.TypedJoiner[testmodels.JoinSummary] = (*testmodels.JoinerJoinSummary)(nil)
}
