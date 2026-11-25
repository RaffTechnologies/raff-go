package spec

import "testing"

// The deprecated VM list status names keep their type and value, so code
// written against the old names compiles and sends the same filter.
func TestListVMsParamsStatusDeprecatedNames(t *testing.T) {
	cases := []struct {
		got  ListVMsParamsStatus
		want string
	}{
		{ListVMsParamsStatusActive, "active"},
		{ListVMsParamsStatusBooting, "booting"},
		{ListVMsParamsStatusFailure, "failure"},
		{ListVMsParamsStatusFinalizing, "finalizing"},
		{ListVMsParamsStatusInitiating, "initiating"},
		{ListVMsParamsStatusPassive, "passive"},
		{ListVMsParamsStatusProvisioning, "provisioning"},
	}
	for _, c := range cases {
		if string(c.got) != c.want || !c.got.Valid() {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}
