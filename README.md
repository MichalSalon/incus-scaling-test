# Incus scaling issue reproducer

## Example

https://github.com/user-attachments/assets/58bc08f6-5bec-443b-8ef8-d29a640810b4

```shell
# Shell 0 - main on the left
./incusTest run alpine-test -d 2 -i 2 -f

# Shell 1 - top right
incus exec alpine-test sh
watch -n 1 -- wget localhost:8080/cpuinfo -q -O -

# Shell 2 - middle right
watch -n 1 -- "cat /sys/fs/cgroup/lxc.payload.alpine-test/cpuset.cpus"

# Shell 3 - bottom right
watch -n 1 -- "incus config show alpine-test | grep limits"
```

---

## Container setup

```shell
# Create default managed network if not present
incus network create lxdbr0

# Start the container
incus launch images:alpine/3.20 alpine-test -s default --network lxdbr0

# Enter the container
incus exec alpine-test sh

# Download the server
wget https://raw.githubusercontent.com/MichalSalon/incus-scaling-test/refs/heads/main/bin/server -O server

# Add the executable bit
chmod +x ./server

# Run the server in foreground or
./server

# Run the server in background (allowing you to exit the container)
nohup ./server &

# Check the output (should contain comma separated list of numbers)
wget localhost:8080/cpuinfo -q -O -
```

## Runner setup

```shell
# Download the binary
curl -L https://raw.githubusercontent.com/MichalSalon/incus-scaling-test/refs/heads/main/bin/incusTest -o incusTest

# Add the executable bit
chmod +x ./incusTest

# Test whether it works
./incusTest run -h

# Run it (alpine-test is the name of the container)
./incusTest run alpine-test -f
```

## Test output

```shell
./incusTest run alpine-test -d 2 -i 2 -f
```

```
Initial thread count 64 [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63]

Setting to 5 cores [22,41,47,53,59] and allowance: 500ms/100ms
✅ Incus     5 cores [22,41,47,53,59]
❌ Container 5 cores [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63]
❌ Cgroup    5 cores [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63]
 - Thread count not equal, attempting manual fix - 
✅ Incus     5 cores [22,41,47,53,59]
✅ Container 5 cores [22,41,47,53,59]
✅ Cgroup    5 cores [22,41,47,53,59]

Setting to 6 cores [11,20,44,46,52,60] and allowance: 600ms/100ms
✅ Incus     6 cores [11,20,44,46,52,60]
❌ Container 6 cores [22,41,47,53,59]
❌ Cgroup    6 cores [22,41,47,53,59]
 - Thread count not equal, attempting manual fix - 
✅ Incus     6 cores [11,20,44,46,52,60]
✅ Container 6 cores [11,20,44,46,52,60]
✅ Cgroup    6 cores [11,20,44,46,52,60]

Setting to 4 cores [0,24,32,37] and allowance: 400ms/100ms
✅ Incus     4 cores [0,24,32,37]
✅ Container 4 cores [0,24,32,37]
✅ Cgroup    4 cores [0,24,32,37]

Setting to 4 cores [3,34,36,38] and allowance: 400ms/100ms
✅ Incus     4 cores [3,34,36,38]
✅ Container 4 cores [3,34,36,38]
✅ Cgroup    4 cores [3,34,36,38]

Setting to 4 cores [5,11,13,36] and allowance: 400ms/100ms
✅ Incus     4 cores [5,11,13,36]
✅ Container 4 cores [5,11,13,36]
✅ Cgroup    4 cores [5,11,13,36]

Setting to 6 cores [26,32,34,36,38,43] and allowance: 600ms/100ms
✅ Incus     6 cores [26,32,34,36,38,43]
❌ Container 6 cores [5,11,13,36]
❌ Cgroup    6 cores [5,11,13,36]
 - Thread count not equal, attempting manual fix - 
✅ Incus     6 cores [26,32,34,36,38,43]
✅ Container 6 cores [26,32,34,36,38,43]
✅ Cgroup    6 cores [26,32,34,36,38,43]

Setting to 2 cores [4,38] and allowance: 200ms/100ms
✅ Incus     2 cores [4,38]
❌ Container 2 cores [26,32,34,36,38,43]
❌ Cgroup    2 cores [26,32,34,36,38,43]
 - Thread count not equal, attempting manual fix - 
✅ Incus     2 cores [4,38]
✅ Container 2 cores [4,38]
✅ Cgroup    2 cores [4,38]

Setting to 9 cores [1,11,15,29,31,44,48,62,63] and allowance: 900ms/100ms
✅ Incus     9 cores [1,11,15,29,31,44,48,62,63]
❌ Container 9 cores [4,38]
❌ Cgroup    9 cores [4,38]
 - Thread count not equal, attempting manual fix - 
✅ Incus     9 cores [1,11,15,29,31,44,48,62,63]
✅ Container 9 cores [1,11,15,29,31,44,48,62,63]
✅ Cgroup    9 cores [1,11,15,29,31,44,48,62,63]
```
