package task

import (
	"testing"

	"github.com/clintjedwards/todo/proto"
	"github.com/google/go-cmp/cmp"
)

func TestToTaskTree(t *testing.T) {
	input := []*proto.Task{
		{
			Id: "parent1",
		},
		{
			Id: "parent2",
		},
		{
			Id:     "parent1 -> child1",
			Parent: "parent1",
		},
		{
			Id:     "parent2 -> child1",
			Parent: "parent2",
		},
		{
			Id:     "parent1 -> child2",
			Parent: "parent1",
		},
		{
			Id:     "parent1 -> child1 -> sub-child1",
			Parent: "parent1 -> child1",
		},
	}

	expectedOutput := map[string]map[string]struct{}{
		"parent1": {
			"parent1 -> child1": {},
			"parent1 -> child2": {},
		},
		"parent2": {
			"parent2 -> child1": {},
		},
		"parent1 -> child1": {
			"parent1 -> child1 -> sub-child1": {},
		},
		"parent1 -> child2":               {},
		"parent2 -> child1":               {},
		"parent1 -> child1 -> sub-child1": {},
	}

	actualOutput, _ := toTaskTree(input)

	if diff := cmp.Diff(expectedOutput, actualOutput); diff != "" {
		t.Errorf("unexpected map values (-want +got):\n%s", diff)
	}
}
