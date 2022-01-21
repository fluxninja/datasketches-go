package sketches

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/base64"
)

var _ = Describe("QuantilesDoublesSketch", func() {
	It("Serializes an empty sketch correctly", func() {
		defaultK := 128
		expectedSerialized := "AQMIBIAAAAA="

		sketch, err := NewDoublesSketch(defaultK)
		Expect(err).To(BeNil())
		serializedBytes, err := sketch.Serialize()
		Expect(err).To(BeNil())
		serializedSketch := base64.StdEncoding.EncodeToString(serializedBytes)
		Expect(serializedSketch).To(Equal(expectedSerialized))
		/*
			TODO
			- empty flag is not set
			- preLongs should be 1 and not 2 for empty sketch
		*/
	})
})
