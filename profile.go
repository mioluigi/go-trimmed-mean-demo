package main

import (
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
)

// startProfiles starts a CPU profile if -cpuprofile is set, and returns a function that stops it and
// writes the memory profile if -memprofile is set (spec FR-11).
func startProfiles(opts options) (func(*slog.Logger), error) {
	var cpuFile *os.File
	if opts.cpuProfile != "" {
		file, err := os.Create(opts.cpuProfile)
		if err != nil {
			return nil, err
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			file.Close()
			return nil, err
		}
		cpuFile = file
	}
	stop := func(logger *slog.Logger) {
		if cpuFile != nil {
			pprof.StopCPUProfile()
			cpuFile.Close()
			logger.Info("wrote CPU profile", "file", opts.cpuProfile)
		}
		if opts.memProfile != "" {
			writeMemoryProfile(opts.memProfile, logger)
		}
	}
	return stop, nil
}

func writeMemoryProfile(path string, logger *slog.Logger) {
	file, err := os.Create(path)
	if err != nil {
		logger.Error("memory profile", "error", err)
		return
	}
	defer file.Close()
	runtime.GC() // report live memory after the run, not garbage waiting for collection
	if err := pprof.WriteHeapProfile(file); err != nil {
		logger.Error("memory profile", "error", err)
		return
	}
	logger.Info("wrote memory profile", "file", path)
}
