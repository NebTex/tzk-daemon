package main

import (
	"os"
)

func (f *Facts) GetContainerStatus() {
	if f.Container != os.Getenv("container") {
		f.HasChanged = true
	}
	f.Container = os.Getenv("container")
}
