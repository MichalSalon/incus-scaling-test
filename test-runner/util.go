package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
)

func ExecuteCommandString(ctx context.Context, command string, args ...string) (string, error) {
	s := &strings.Builder{}
	exitCode, err := ExecuteCommand(ctx, s, command, args...)
	if err != nil {
		return "", err
	}
	if exitCode > 0 {
		return "", fmt.Errorf("unexpected error exitCode: %d", exitCode)
	}
	return strings.TrimSpace(s.String()), nil
}

func ExecuteCommand(ctx context.Context, output io.Writer, command string, args ...string) (int, error) {
	// fmt.Printf("Executing: %s %s\n", command, strings.Join(args, " "))
	c := exec.CommandContext(ctx, command, args...)
	c.Env = os.Environ()
	c.Stdout = output
	c.Stderr = output
	if err := c.Start(); err != nil {
		return 0, err
	}

	if waitErr := c.Wait(); waitErr != nil {
		waitErr = fmt.Errorf("exec error: %w", waitErr)

		var exitError *exec.ExitError
		if errors.As(waitErr, &exitError) {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), waitErr
			}
		}
		return c.ProcessState.ExitCode(), waitErr
	}
	return c.ProcessState.ExitCode(), nil
}

func GetThreads(threads int, maxThread int) []int {
	out := make([]int, threads)
	duplicityCheck := make(map[int]struct{}, threads)
	for i := range out {
		for {
			num := rand.Intn(maxThread) // no + 1 as maxThread is already offset by 1
			if _, exists := duplicityCheck[num]; exists {
				continue
			}
			duplicityCheck[num] = struct{}{}
			out[i] = num
			break
		}
	}
	slices.Sort(out)
	return out
}

func ParseCpuSet(cpuSet string) ([]int, error) {
	out := make([]int, 0, runtime.NumCPU())

	// cores may be in a simple 0,2,3,4,5,9,10,15,18 format or in 0,2-5,9-10,15,18
	coreSets := strings.Split(cpuSet, ",")
	for _, coreSet := range coreSets {
		coreRange := strings.Split(coreSet, "-")
		if len(coreRange) == 1 {
			i, err := strconv.Atoi(coreSet)
			if err != nil {
				return nil, err
			}
			out = append(out, i)
			continue
		}

		start, err := strconv.Atoi(coreRange[0])
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(coreRange[1])
		if err != nil {
			return nil, err
		}

		for i := start; i <= end; i++ {
			out = append(out, i)
		}
	}
	slices.Sort(out)
	return out, nil
}

func GetCpuLimitAndAllowance(threads []int) (limit string, allowance string) {
	const cfsPeriodOneCpu = 100

	cores := float64(len(threads))
	switch cores {
	case 0:
		lastThread := runtime.NumCPU() - 1
		limit = fmt.Sprintf("%d-%d", lastThread, lastThread)
	case 1:
		limit = fmt.Sprintf("%d-%d", threads[0], threads[0])
	default:
		b := strings.Builder{}
		for i, v := range threads {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.FormatInt(int64(v), 10))
		}
		limit = b.String()
	}

	allowance = fmt.Sprintf("%dms/%dms", int64(math.Round(cores*cfsPeriodOneCpu)), cfsPeriodOneCpu)
	return limit, allowance
}

func intArrToString(intArray []int) string {
	stringArray := make([]string, len(intArray))
	for i, num := range intArray {
		stringArray[i] = strconv.Itoa(num)
	}
	return strings.Join(stringArray, ",")
}
