package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDOB(t *testing.T) {
	t.Run("Valid Age", func(t *testing.T) {
		assert.True(t, validateDOB("02/03/1990"), "Expected true")
		assert.True(t, validateDOB("07/03/1825"), "Expected true")
	})

	t.Run("Wrong Date", func(t *testing.T) {
		assert.False(t, validateDOB("02/31/2000"), "Expected false")
	})

	t.Run("Date too far", func(t *testing.T) {
		assert.False(t, validateDOB("02/20/0000"), "Expected false")
	})

	t.Run("Age below 18", func(t *testing.T) {
		assert.False(t, validateDOB("02/20/2022"), "Expected false")
	})

	t.Run("Age above 200", func(t *testing.T) {
		assert.False(t, validateDOB("02/20/1824"), "Expected false")
	})
}
