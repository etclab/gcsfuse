// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Write out a file of a certain size, close it, then measure the performance
// of doing the following:
//
// 1.  Open the file.
// 2.  Read it from start to end with a configurable buffer size.
package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/googlecloudplatform/gcsfuse/v2/benchmarks/internal/format"
	"github.com/googlecloudplatform/gcsfuse/v2/benchmarks/internal/percentile"
)

var fDir = flag.String("dir", "", "Directory within which to write the file.")
var fMountCmd = flag.String("mount_cmd", "", "Command to remount bucket. If not passed, never unmounts bucket.")
var fDuration = flag.Duration("duration", 10*time.Second, "How long to run.")
var fFileSize = flag.Int64("file_size", 1<<26, "Size of file to use.")
var fReadSize = flag.Int64("read_size", 1<<20, "Size of each call to read(2).")
var fRawOut = flag.String("raw_out", "", "File to which to append raw results in bytes/sec.")

////////////////////////////////////////////////////////////////////////
// main logic
////////////////////////////////////////////////////////////////////////

func mountBucket() (err error) {
	if *fMountCmd == "" {
		return
	}

	err = exec.Command(*fMountCmd).Run()
	return
}

func umountBucket() (err error) {
	if *fMountCmd == "" {
		return
	}

	err = exec.Command("fusermount", "-u", *fDir).Run()
	return
}

func run() (err error) {
	if *fDir == "" {
		err = errors.New("You must set --dir.")
		return
	}

	umountBucket()
	err = mountBucket()
	if err != nil {
		err = fmt.Errorf("mountBucket: %w", err)
		return
	}

	// Create a temporary file.
	log.Printf("Creating a temporary file in %s.", *fDir)

	f, err := os.CreateTemp(*fDir, "sequential_read")
	if err != nil {
		err = fmt.Errorf("TempFile: %w", err)
		return
	}

	path := f.Name()

	// Make sure we clean it up later.
	defer func() {
		log.Printf("Deleting %s.", path)
		mountBucket()
		os.Remove(path)
		umountBucket()
	}()

	// Fill it with random content.
	log.Printf("Writing %d random bytes.", *fFileSize)
	_, err = io.Copy(f, io.LimitReader(rand.Reader, *fFileSize))
	if err != nil {
		err = fmt.Errorf("Copying random bytes: %w", err)
		return
	}

	// Finish off the file.
	err = f.Close()
	if err != nil {
		err = fmt.Errorf("Closing file: %w", err)
		return
	}

	err = umountBucket()
	if err != nil {
		err = fmt.Errorf("umountBucket: %w", err)
		return
	}

	// Run several iterations.
	log.Printf("Measuring for %v...", *fDuration)

	var fullFileRead percentile.DurationSlice
	var singleOpenCall percentile.DurationSlice
	var singleReadCall percentile.DurationSlice
	var fullReadLoop percentile.DurationSlice
	var singleCloseCall percentile.DurationSlice
	buf := make([]byte, *fReadSize)

	overallStartTime := time.Now()
	for len(fullFileRead) == 0 || time.Since(overallStartTime) < *fDuration {
		err = mountBucket()
		if err != nil {
			err = fmt.Errorf("mountBucket: %w", err)
			return
		}

		fileStartTime := time.Now()

		f, err = os.Open(path)
		singleOpenCall = append(singleOpenCall, time.Since(fileStartTime))
		if err != nil {
			err = fmt.Errorf("Opening file: %w", err)
			return
		}

		readLoopStartTime := time.Now()
		var readErr error
		for readErr == nil {
			readStartTime := time.Now()
			_, readErr = f.Read(buf)
			singleReadCall = append(singleReadCall, time.Since(readStartTime))
		}
		fullReadLoop = append(fullReadLoop, time.Since(readLoopStartTime))

		closeStartTime := time.Now()
		err = f.Close()

		endTime := time.Now()
		singleCloseCall = append(singleCloseCall, endTime.Sub(closeStartTime))
		fullFileRead = append(fullFileRead, endTime.Sub(fileStartTime))

		if err != nil {
			err = fmt.Errorf("Closing file after reading: %w", err)
			return
		}

		switch {
		case readErr == io.EOF:
			readErr = nil

		case readErr != nil:
			readErr = fmt.Errorf("Reading: %w", readErr)
			return
		}

		err = umountBucket()
		if err != nil {
			err = fmt.Errorf("umountBucket: %w", err)
			return
		}
	}

	sort.Sort(fullFileRead)
	sort.Sort(singleOpenCall)
	sort.Sort(singleReadCall)
	sort.Sort(fullReadLoop)
	sort.Sort(singleCloseCall)

	log.Printf(
		"Read the file %d times, using %d calls to read(2).",
		len(fullFileRead),
		len(singleReadCall))

	// Report.
	ptiles := []int{50, 90, 98}

	reportSlice := func(
		name string,
		bytesPerObservation int64,
		observations percentile.DurationSlice) {
		fmt.Printf("\n%s:\n", name)
		for _, ptile := range ptiles {
			d := percentile.Duration(observations, ptile)
			seconds := float64(d) / float64(time.Second)
			bandwidthBytesPerSec := float64(bytesPerObservation) / seconds

			fmt.Printf(
				"  %02dth ptile: %10v (%s/s)\n",
				ptile,
				d,
				format.Bytes(bandwidthBytesPerSec))
		}
	}

	reportSlice("Full-file read times", *fFileSize, fullFileRead)
	reportSlice("open call latencies", 0, singleOpenCall)
	reportSlice("read(2) single call latencies", *fReadSize, singleReadCall)
	reportSlice("read loop times", *fFileSize, fullReadLoop)
	reportSlice("close call latencies", 0, singleCloseCall)

	fmt.Println()

	if *fRawOut == "" {
		return
	}

	f, err = os.OpenFile(*fRawOut, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = fmt.Errorf("open raw out: %w", err)
		return
	}

	for _, dur := range fullFileRead {
		seconds := float64(dur) / float64(time.Second)
		bandwidthBytesPerSec := float64(*fFileSize) / seconds
		_, err = fmt.Fprintf(f, "%f\n", bandwidthBytesPerSec)
		if err != nil {
			return
		}
	}

	f.Close()
	return
}

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()

	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
