package pkg

import "testing"

func TestHaveController(t *testing.T) {
	testCases := []struct {
		Controllers string
		Controller  string
		isContained bool
	}{
		{
			Controllers: "*",
			Controller:  "deployment",
			isContained: true,
		},
		{
			Controllers: "*,-deployment",
			Controller:  "deployment",
			isContained: false,
		},
		{
			Controllers: "deployment",
			Controller:  "job",
			isContained: false,
		},
		{
			Controllers: "*,-job",
			Controller:  "job",
			isContained: false,
		},
	}
	for _, tc := range testCases {
		res := HaveController(tc.Controllers, tc.Controller)
		if res != tc.isContained {
			t.Fail()
		}
	}
}
