//go:build !darwin

package main

func readMacHardwareUUID() string { return "" }
