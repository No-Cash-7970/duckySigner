// Copyright (C) 2019-2024 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

// XXX: Modified from https://github.com/algorand/go-algorand/tree/8b6c443d6884b4c0d3e3b3faf35b886fb81598a3/util/codecs/json_test.go

package codecs_test

import (
	"bytes"
	"hash/fnv"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	. "duckysigner/kmd/codecs"
)

type testValue struct {
	Bool   bool
	String string
	Int    int
}

var _ = Describe("JSON Codecs", func() {
	It("has working isDefaultValue()", func() {
		t := GinkgoT()

		PartitionTest(t)
		t.Parallel()

		a := require.New(t)

		v := testValue{
			Bool:   true,
			String: "default",
			Int:    1,
		}
		def := testValue{
			Bool:   true,
			String: "default",
			Int:    2,
		}

		objectValues := CreateValueMap(v)
		defaultValues := CreateValueMap(def)

		a.True(IsDefaultValue("Bool", objectValues, defaultValues))
		a.True(IsDefaultValue("String", objectValues, defaultValues))
		a.False(IsDefaultValue("Int", objectValues, defaultValues))
		a.True(IsDefaultValue("Missing", objectValues, defaultValues))
	})

	It("has working SaveObjectToFile()", func() {
		t := GinkgoT()

		PartitionTest(t)
		t.Parallel()

		type TestType struct {
			A uint64
			B string
		}

		obj := TestType{1024, "test"}

		// prettyFormat = false
		{
			filename := path.Join(t.TempDir(), "test.json")
			SaveObjectToFile(filename, obj, false)
			data, err := os.ReadFile(filename)
			require.NoError(t, err)
			expected := `{"A":1024,"B":"test"}
`
			require.Equal(t, expected, string(data))
		}

		// prettyFormat = true
		{
			filename := path.Join(t.TempDir(), "test.json")
			SaveObjectToFile(filename, obj, true)
			data, err := os.ReadFile(filename)
			require.NoError(t, err)
			expected := "{\n\t\"A\": 1024,\n\t\"B\": \"test\"\n}\n"
			require.Equal(t, expected, string(data))
		}
	})

	Describe("WriteNonDefaultValue()", func() {
		type TestType struct {
			Version       uint32
			Archival      bool
			GossipFanout  int
			NetAddress    string
			ReconnectTime time.Duration
		}

		defaultObject := TestType{
			Version:       1,
			Archival:      true,
			GossipFanout:  50,
			NetAddress:    "Denver",
			ReconnectTime: 60 * time.Second,
		}

		DescribeTable("In various scenarios",
			func(in TestType, out string, ignore []string) {
				var writer bytes.Buffer
				err := WriteNonDefaultValues(&writer, in, defaultObject, ignore)
				Expect(err).NotTo(HaveOccurred())
				Expect(writer.String()).To(Equal(out))
			},
			Entry("works with all defaults", defaultObject, "{\n}", []string{}),
			Entry("works with some defaults",
				TestType{
					Version:       1,
					Archival:      false,
					GossipFanout:  25,
					NetAddress:    "Denver",
					ReconnectTime: 60 * time.Nanosecond,
				},
				"{\n\t\"Archival\": false,\n\t\"GossipFanout\": 25,\n\t\"ReconnectTime\": 60\n}",
				[]string{},
			),
			Entry("works with ignore",
				defaultObject,
				"{\n\t\"Version\": 1,\n\t\"Archival\": true,\n\t\"GossipFanout\": 50,\n\t\"NetAddress\": \"Denver\",\n\t\"ReconnectTime\": 60000000000\n}",
				[]string{"Version", "Archival", "GossipFanout", "NetAddress", "ReconnectTime"},
			),
		)
	})
})

// XXX: The following is from https://github.com/algorand/go-algorand/tree/8b6c443d6884b4c0d3e3b3faf35b886fb81598a3/test/partitiontest/filtering.go

// PartitionTest checks if the current partition should run this test, and skips it if not.
func PartitionTest(t FullGinkgoTInterface) {
	pt, found := os.LookupEnv("PARTITION_TOTAL")
	if !found {
		return
	}
	partitions, err := strconv.Atoi(pt)
	if err != nil {
		return
	}
	pid := os.Getenv("PARTITION_ID")
	partitionID, err := strconv.Atoi(pid)
	if err != nil {
		return
	}
	name := t.Name()
	_, file, _, _ := runtime.Caller(1) // get filename of caller to PartitionTest
	nameNumber := stringToUint64(file + ":" + name)
	idx := nameNumber % uint64(partitions)
	if idx != uint64(partitionID) {
		t.Skipf("skipping %s due to partitioning: assigned to %d but I am %d of %d", name, idx, partitionID, partitions)
	}
}

func stringToUint64(str string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}
