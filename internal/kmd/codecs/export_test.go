// This file is for exporting private functions and variables only to tests.
// This allows for a private function/variable to remain private outside of testing.
// A trick borrowed from https://stackoverflow.com/a/60813569 and
// https://medium.com/@robiplus/golang-trick-export-for-test-aa16cbd7b8cd

package codecs

var CreateValueMap = createValueMap
var IsDefaultValue = isDefaultValue
