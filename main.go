package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type FileTimestamps struct {
	Atime time.Time
	Mtime time.Time
}

func registerDir(dir string) (map[string]FileTimestamps, error) {
	timestamps := make(map[string]FileTimestamps)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: cannot access %s: %v\n", path, err)
			return nil
		}

		stat := info.Sys().(*syscall.Stat_t)
		ts := FileTimestamps{
			Atime: time.Unix(stat.Atim.Sec, stat.Atim.Nsec),
			Mtime: info.ModTime(),
		}
		timestamps[path] = ts
		return nil
	})

	return timestamps, err
}

func executeCommand(command []string, workDir string) error {
	cmd := exec.Command(command[0], command[1:]...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func resetTimestamps(dir string, defaultStamp time.Time, timestamps map[string]FileTimestamps) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: cannot access %s: %v\n", path, err)
			return nil
		}

		stat := info.Sys().(*syscall.Stat_t)
		curAtime := time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
		curMtime := info.ModTime()

		originalTs, exist := timestamps[path]
		if !exist {
			// new file. let's set to 0 or SOURCE_DATE_EPOCH
			originalTs.Atime = defaultStamp
			originalTs.Mtime = defaultStamp
			log.Println("found new file:", path)
		} else {
			if curAtime.Equal(originalTs.Atime) && curMtime.Equal(originalTs.Mtime) {
				return nil
			}
			log.Println("fixing timestamp for:", path)
		}
		atimeSpec := unix.Timespec{Sec: originalTs.Atime.Unix(), Nsec: int64(originalTs.Atime.Nanosecond())}
		mtimeSpec := unix.Timespec{Sec: originalTs.Mtime.Unix(), Nsec: int64(originalTs.Mtime.Nanosecond())}

		// Use UtimesNanoAt with AT_SYMLINK_NOFOLLOW flag
		return unix.UtimesNanoAt(-100, path, []unix.Timespec{atimeSpec, mtimeSpec}, unix.AT_SYMLINK_NOFOLLOW)
	})

	return err
}

func main() {
	changeDir := flag.Bool("cd", false, "change to <directory> before executing command")
	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatalf("Usage: %s <directory> <command> [args ...]\n", os.Args[0])
	}

	dir := flag.Args()[0]
	command := flag.Args()[1:]
	workDir := ""
	if *changeDir {
		workDir = dir
	}

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		log.Fatalf("Error: %s is not a valid directory: %v\n", dir, err)
	}

	// Step 1: Scan original timestamps
	log.Print("Scanning original timestamps...")
	originalTimestamps, err := registerDir(dir)
	if err != nil {
		log.Fatalf("Error scanning directory: %v\n", err)
	}
	log.Printf("Found %d files/directories\n", len(originalTimestamps))

	// Step 2: Execute command
	log.Printf("Executing command: %s\n", command)
	if err := executeCommand(command, workDir); err != nil {
		log.Fatalf("Command failed: %v\n", err)
	}

	// Step 3: Reset timestamps
	log.Print("Resetting timestamps...")

	defaultTimestamp := time.Unix(0, 0)
	// set default timestamp for new files to SDE if specified
	sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH")
	if sourceDateEpoch != "" {
		timestamp, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
		if err == nil {
			defaultTimestamp = time.Unix(timestamp, 0)
		}
	}

	if err := resetTimestamps(dir, defaultTimestamp, originalTimestamps); err != nil {
		log.Fatalf("Error resetting timestamps: %v\n", err)
	}
}
