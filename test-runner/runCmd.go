package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	runCmd.PersistentFlags().IntP("max-cores", "c", runtime.NumCPU(), "Max cpu cores to try to scale to.")
	runCmd.PersistentFlags().IntP("check-delay", "d", 5, "Delay between incus command and validation in seconds.")
	runCmd.PersistentFlags().IntP("iteration-delay", "i", 5, "Delay between iterations in seconds.")
	runCmd.PersistentFlags().BoolP("enable-fixup", "f", false, "Whether the fixup by writing directly to cgroup should be attempted.")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the test",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]
		if containerName == "" {
			return errors.New("container name is required")
		}

		maxCores, err := cmd.Flags().GetInt("max-cores")
		if err != nil {
			return err
		}

		checkDelay, err := cmd.Flags().GetInt("check-delay")
		if err != nil {
			return err
		}

		iterationDelay, err := cmd.Flags().GetInt("iteration-delay")
		if err != nil {
			return err
		}

		enableFixup, err := cmd.Flags().GetBool("enable-fixup")
		if err != nil {
			return err
		}

		initialCores, err := getCoreCount(cmd.Context(), containerName)
		if err != nil {
			return err
		}
		fmt.Printf("Initial thread count %d [%s]\n\n", len(initialCores), intArrToString(initialCores))

		for {
			// get up to 10 different threads in range [0,maxCores)
			threadCount := rand.Intn(min(10, maxCores)) + 1
			threads := GetThreads(threadCount, maxCores)
			cores, allowance := GetCpuLimitAndAllowance(threads)

			fmt.Printf("Setting to %d cores [%s] and allowance: %s\n", threadCount, cores, allowance)

			exitCode, err := ExecuteCommand(
				cmd.Context(),
				os.Stdout,
				"incus",
				"config",
				"set",
				containerName,
				fmt.Sprintf(`limits.cpu=%s`, cores),
				fmt.Sprintf(`limits.cpu.allowance=%s`, allowance),
			)
			if err != nil {
				return err
			}
			if exitCode > 0 {
				return fmt.Errorf("unexpected error exitCode: %d", exitCode)
			}

			time.Sleep(time.Second * time.Duration(checkDelay))

			areEqual, err := checkCores(cmd.Context(), containerName, threads)
			if err != nil {
				return err
			}

			if enableFixup && !areEqual {
				fmt.Println(" - Thread count not equal, attempting manual fix - ")
				if err := os.WriteFile("/sys/fs/cgroup/lxc.payload."+containerName+"/cpuset.cpus", []byte(cores), 0o600); err != nil {
					return err
				}
				_, err := checkCores(cmd.Context(), containerName, threads)
				if err != nil {
					return err
				}
			}

			fmt.Println("")
			time.Sleep(time.Second * time.Duration(iterationDelay))
		}
	},
}

func checkCores(ctx context.Context, containerName string, threads []int) (bool, error) {
	incusCpuSet, err := ExecuteCommandString(ctx, "incus", "config", "get", containerName, "limits.cpu")
	if err != nil {
		return false, err
	}
	incusCores, err := ParseCpuSet(incusCpuSet)
	if err != nil {
		return false, err
	}
	slicesEqualPrint("Incus    ", threads, incusCores)

	actualThreads, err := getCoreCount(ctx, containerName)
	if err != nil {
		return false, err
	}
	slicesEqualPrint("Container", threads, actualThreads)

	cgroupCpuSet, err := os.ReadFile("/sys/fs/cgroup/lxc.payload." + containerName + "/cpuset.cpus")
	if err != nil {
		return false, err
	}
	cgroupCores, err := ParseCpuSet(strings.TrimSpace(string(cgroupCpuSet)))
	if err != nil {
		return false, err
	}
	slicesEqualPrint("Cgroup   ", threads, cgroupCores)

	return slices.Equal(threads, actualThreads), nil
}

func getCoreCount(ctx context.Context, containerName string) ([]int, error) {
	s, err := ExecuteCommandString(ctx, "incus", "exec", containerName, "--", "wget", "localhost:8080/cpuinfo", "-q", "-O", "-")
	if err != nil {
		return nil, err
	}

	coresStr := strings.Split(s, ",")
	cores := make([]int, len(coresStr))
	for i := range cores {
		cores[i], _ = strconv.Atoi(coresStr[i])
	}
	return cores, nil
}

func slicesEqualPrint(name string, a, b []int) {
	status := "❌" // ❎
	if slices.Equal(a, b) {
		status = "✅"
	}
	fmt.Printf("%s %s %d cores [%s]\n", status, name, len(a), intArrToString(b))
}
