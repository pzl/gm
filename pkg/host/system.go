package host

import (
	"context"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"

	"net/http"

	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/pzl/manager/pkg/config"

	"github.com/coreos/go-systemd/dbus"
	"github.com/sirupsen/logrus"
)

type Versions struct {
	Linux  string `json:"linux"`
	Podman string `json:"podman"`
}

func RegisterSystemHandlers(serveMux *http.ServeMux, ctx context.Context) {
	log := ctx.Value(config.LogKey).(*logrus.Logger)
	c := ctx.Value(config.DbusKey).(*dbus.Conn)
	serveMux.HandleFunc("/api/system/reload", func(w http.ResponseWriter, r *http.Request) {
		log.Warn("reloading systemd")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := ReloadSystemD(c)
		if err != nil {
			log.WithError(err).Error("error reloading systemd")
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte("reloaded"))
		}
	})

	serveMux.HandleFunc("/api/system/versions/", func(w http.ResponseWriter, r *http.Request) {
		v := Versions{
			Linux:  Uname(),
			Podman: PodmanVersion(),
		}

		js, err := json.Marshal(v)
		if err != nil {
			log.WithError(err).Error("error encoding version info to json")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	serveMux.HandleFunc("/api/system/vpn/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		out, err := exec.Command("systemctl", "is-active", "vpn-out.service").Output()
		if err != nil && len(out) == 0 {
			log.WithError(err).Error("error checking if vpn-out.service is running")
			w.Write([]byte(`{"error": "` + err.Error() + `"}`))
			return
		}
		if strings.TrimSpace(string(out)) == "active" {
			w.Write([]byte("true"))
			return
		}
		w.Write([]byte("false"))
	})

	serveMux.HandleFunc("/api/system/memory/", func(w http.ResponseWriter, r *http.Request) {

		m := RAM()
		if m == nil {
			http.Error(w, "dunno", http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(m)
		if err != nil {
			log.WithError(err).Error("error json marshaling memory info")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}

/*
 * re-scans unit files. Effectively `systemd daemon-reload`
 */
func ReloadSystemD(c *dbus.Conn) error {
	return c.Reload()
}

/*
 * Gets local podman version as a string
 */
func PodmanVersion() string {
	out, err := exec.Command("podman", "version").Output()
	if err != nil {
		return err.Error()
	}
	lines := bytes.Split(out, []byte{'\n'})
	fields := strings.Fields(string(lines[0]))
	return fields[1]
}

type Meminfo struct {
	Total int64 `json:"total"`
	Avail int64 `json:"avail"`
}

/*
 * Returns total and available RAM
 * in kB
 */
func RAM() *Meminfo {
	m := Meminfo{}
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil
	}
	defer f.Close()

	buf := make([]byte, 512)
	if _, err := f.Read(buf); err != nil {
		return nil
	}

	lines := bytes.Split(buf, []byte{'\n'})
	t := strings.Fields(string(lines[0]))
	av := strings.Fields(string(lines[2]))
	m.Total, _ = strconv.ParseInt(t[1], 10, 64)
	m.Avail, _ = strconv.ParseInt(av[1], 10, 64)

	return &m
}

/*
 * Returns Kernel version as a string
 */
func Uname() string {
	u := syscall.Utsname{}
	if err := syscall.Uname(&u); err != nil {
		return err.Error()
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

func blockDevices() ([]*BlockDevice, error) {
	var b []*BlockDevice

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

		b = append(b, &BlockDevice{
			maj:       int(maj),
			min:       int(min),
			RawSize:   uint64(size),
			Block:     fields[3], // e.g.  sda
			Removable: remove,
		})
	}
	return b, nil
}

func mountInfo(b []*BlockDevice) {
	mounts, err := ioutil.ReadFile("/proc/self/mounts")
	if err != nil {
		return
	}

	lines := strings.Split(string(mounts), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		for _, block := range b {
			if block.Mount != "" {
				continue // if we have already identified a device, skip it
			}
			if fields[0] == "/dev/"+block.Block {
				block.Mount = fields[1]
				block.Type = fields[2]
				break
			}
		}
	}
}

func sizeInfo(b *BlockDevice) {
	if b.Mount == "" {
		return
	}
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(b.Mount, &fs)
	if err != nil {
		return
	}

	reserved := (fs.Bfree - fs.Bavail) * uint64(fs.Bsize)

	b.All = fs.Blocks*uint64(fs.Bsize) - reserved
	b.Free = fs.Bavail * uint64(fs.Bsize)
	b.TInodes = fs.Files
	b.FInodes = fs.Ffree
	b.Used = b.All - b.Free

	return
}

func DiskInfo() []*BlockDevice {
	blocks, err := blockDevices()
	if err != nil {
		return nil
	}
	mountInfo(blocks)

	for _, b := range blocks {
		sizeInfo(b)
	}

	return blocks
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
