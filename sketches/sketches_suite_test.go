package sketches

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSketches(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sketches Suite")
}
