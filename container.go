package main

import (
	"os"
)

//GetContainerStatus check if the current node is running on lxc or not
func (f *Facts) GetContainerStatus() {
	if f.Container != os.Getenv("container") {
		f.HasChanged = true
	}
	f.Container = os.Getenv("container")
}
