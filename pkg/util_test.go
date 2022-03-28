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

func TestIsDeployByPod(t *testing.T) {
	testCases := []struct {
		Controllers   string
		isDeployByPod bool
	}{
		{Controllers: "*,-deployment", isDeployByPod: true},
		{Controllers: "*,-job", isDeployByPod: true},
		{Controllers: "deployment,job", isDeployByPod: true},
		{Controllers: "deployment,job,replicaset", isDeployByPod: false},
		{Controllers: "*", isDeployByPod: false},
	}
	for _, tc := range testCases {
		if IfDeployByPod(tc.Controllers) != tc.isDeployByPod {
			t.Fatal(tc.Controllers)
		}
	}

}
