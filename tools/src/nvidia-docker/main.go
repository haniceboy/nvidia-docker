// Copyright (c) 2015, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"log"
	"os"
	"strings"

	"docker"
	"nvidia"
)

const PluginName = "nvidia"

const (
	EnvDockerBin = "NV_DOCKER_BIN"
	EnvGPU       = "NV_GPU"
)

var GPU []string

func init() {
	log.SetPrefix(os.Args[0] + " | ")

	GPU = strings.FieldsFunc(os.Getenv(EnvGPU), func(c rune) bool {
		return c == ' ' || c == ','
	})
	bin := strings.Fields(os.Getenv(EnvDockerBin))
	docker.SetBinary(bin...)
}

func assert(err error) {
	if err != nil {
		log.Panicln("Error:", err)
	}
}

func exit() {
	code := 0
	if recover() != nil {
		code = 1
	}
	os.Exit(code)
}

func Setup(image string) []string {
	vols, err := volumesNeeded(image)
	assert(err)
	if vols == nil {
		return nil
	}
	assert(cudaIsSupported(image))

	// FIXME avoid looking up every devices
	devs, err := nvidia.LookupDevices()
	assert(err)
	d, err := devicesArgs(devs)
	assert(err)
	v, err := volumesArgs(vols)
	assert(err)

	return append(d, v...)
}

func main() {
	var image string

	args := os.Args[1:]
	defer exit()

	assert(nvidia.Init())
	defer func() { assert(nvidia.Shutdown()) }()

	command, i, err := docker.ParseArgs(args)
	assert(err)
	if command != "" {
		image, i, err = docker.ParseArgs(args[i+1:], command)
		assert(err)
	}
	switch command {
	case "create":
		fallthrough
	case "run":
		if image != "" {
			nvargs := Setup(image)
			args = append(args[:i], append(nvargs, args[i:]...)...)
		}
	default:
	}

	assert(nvidia.LoadUVM())
	assert(docker.Docker(args...))
}