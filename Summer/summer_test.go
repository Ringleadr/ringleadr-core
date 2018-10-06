package Summer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSum(t *testing.T) {
	output := Sum(1,2,3)
	assert.Equal(t, 6, output, "The sum of 1, 2 and 3 should equal 6")
}

func TestSumNoInput(t *testing.T) {
	output := Sum()
	assert.Equal(t, 0, output ,"The sum of no numbers should equal 0")
}
