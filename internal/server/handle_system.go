package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	jsoniter "github.com/json-iterator/go"
	"github.com/pzl/gm/internal/podman"
)

func (s *Server) SystemGetMemory(w http.ResponseWriter, r *http.Request) {
	ram, err := getRAM()
	if err != nil {
		s.Log.WithError(err).Error("error fetching RAM")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	_ = jsCfg.NewEncoder(w).Encode(ram)
}

func (s *Server) SystemGetVersions(w http.ResponseWriter, r *http.Request) {
	podinfo, err := s.Pod.Info(r.Context())
	if err != nil {
		s.Log.WithError(err).Error("error fetching podman info")
	}
	v := struct {
		Linux  string              `json:"linux"`
		Podman jsoniter.RawMessage `json:"podman"`
	}{
		Linux:  kernelVer(),
		Podman: podinfo,
	}
	jsCfg.NewEncoder(w).Encode(v)
}

func (s *Server) SystemReloadSysD(w http.ResponseWriter, r *http.Request) {
	err := s.DBus.Reload()
	status := "success"
	message := "success"
	if err != nil {
		status = "error"
		message = err.Error()
	}
	_, _ = w.Write([]byte(`{"status":"` + status + `", "message": "` + message + `"}`))
}

func (s *Server) SystemGetDisk(w http.ResponseWriter, r *http.Request) {
	blocks, err := blockDevices()
	if err != nil {
		s.Log.WithError(err).Error("error fetching disk info")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	// add mount info
	if mounts, err := ioutil.ReadFile("/proc/self/mounts"); err == nil {
		lines := strings.Split(string(mounts), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			for i := range blocks {
				if blocks[i].Mount != "" {
					continue // if we have already identified a device, skip it
				}
				if fields[0] == "/dev/"+blocks[i].Block {
					// add Mount info
					blocks[i].Mount = fields[1]
					blocks[i].Type = fields[2]

					// add size info
					fs := syscall.Statfs_t{}
					if err := syscall.Statfs(blocks[i].Mount, &fs); err == nil {
						reserved := (fs.Bfree - fs.Bavail) * uint64(fs.Bsize)

						blocks[i].All = fs.Blocks*uint64(fs.Bsize) - reserved
						blocks[i].Free = fs.Bavail * uint64(fs.Bsize)
						blocks[i].TInodes = fs.Files
						blocks[i].FInodes = fs.Ffree
						blocks[i].Used = blocks[i].All - blocks[i].Free
					}

					break
				}
			}
		}
	}

	jsCfg.NewEncoder(w).Encode(blocks)
}

/*
 * Returns total and available RAM
 * in kB
 */
func getRAM() (map[string]int64, error) {
	m := make(map[string]int64, 2)
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return m, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	if _, err := f.Read(buf); err != nil {
		return m, err
	}

	lines := bytes.Split(buf, []byte{'\n'})
	t := strings.Fields(string(lines[0]))
	av := strings.Fields(string(lines[2]))
	m["total"], _ = strconv.ParseInt(t[1], 10, 64)
	m["avail"], _ = strconv.ParseInt(av[1], 10, 64)

	return m, nil
}

/*
 * Returns Kernel version as a string
 */
func kernelVer() string {
	u := syscall.Utsname{}
	if err := syscall.Uname(&u); err != nil {
		return "unknown: " + err.Error()
	}
	return charsToString(u.Release)
}

// required for crappy Utsname struct converting
func charsToString(ca [65]int8) string {
	s := make([]byte, len(ca))
	var lens int
	for lens = 0; lens < len(ca); lens++ {
		if ca[lens] == 0 {
			break
		}
		s[lens] = uint8(ca[lens])
	}
	return string(s[0:lens])
}

/*
 * Gets local podman version as a string
 * deprecated
 */
func podmanVersion(p podman.Conn) string {
	// unused, in favor of REST API

	out, err := exec.Command("podman", "version").Output()
	if err != nil {
		return "unknown: " + err.Error()
	}
	lines := bytes.Split(out, []byte{'\n'})
	fields := strings.Fields(string(lines[0]))
	return fields[1]
}

// Disk / Block stuff

type BlockDevice struct {
	// populated from /proc/partitions
	maj       int
	min       int
	RawSize   uint64
	Block     string
	Removable bool

	// filled by mountInfo
	Mount string
	Type  string

	// filled by sizeInfo
	All     uint64
	Used    uint64
	Free    uint64
	TInodes uint64
	FInodes uint64
}

/*
	Mount detection.
	/sys/block/sd*        for blocks by name
	/sys/dev/block/*      for blocks by maj:min ID

	/proc/self/mounts     for blockname<->mount                (aka /proc/mounts, /etc/mtab)
	/proc/self/mountinfo  for maj:min ID/blockname <-> mount

	/proc/partitions      for maj:min <-> blockname + size

	/sys/block/sdX/removable  1/0 if block is removable


*/
func blockDevices() ([]BlockDevice, error) {
	var b []BlockDevice

	partitions, err := ioutil.ReadFile("/proc/partitions")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(partitions), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if fields[0] == "major" {
			continue
		}

		var remove bool
		rem, err := ioutil.ReadFile("/sys/block/" + fields[3] + "/removable")
		if err != nil {
			remove = false
		} else {
			remove, _ = strconv.ParseBool(string(rem))
		}

		maj, _ := strconv.ParseInt(fields[0], 10, 32)
		min, _ := strconv.ParseInt(fields[1], 10, 32)
		size, _ := strconv.ParseInt(fields[2], 10, 64)

		b = append(b, BlockDevice{
			maj:       int(maj),
			min:       int(min),
			RawSize:   uint64(size),
			Block:     fields[3], // e.g.  sda
			Removable: remove,
		})
	}
	return b, nil
}
