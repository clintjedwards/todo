package tui

import "github.com/clintjedwards/todo/proto"

type detailView struct {
	selectedTask *proto.Task
}

func (v *detailView) display(m model) string {
	return ""
}

func newDetailView() (*detailView, error) {
	return nil, nil
}
