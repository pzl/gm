package main

import (
	"syscall"
	"io/ioutil"
	"strconv"
	"strings"

	"net/http"

	"os"
	"os/exec"
	"bytes"
	"encoding/json"

	"github.com/coreos/go-systemd/dbus"
)

type Versions struct {
	Linux string `json:"linux"`
	Rkt   string `json:"rkt"`
}

func RegisterSystemHandlers(serveMux *http.ServeMux, c *dbus.Conn) {
	serveMux.HandleFunc("/api/system/reload", func (w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "*")
		err := ReloadSystemD(c)
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte("reloaded"))
		}
	})

	serveMux.HandleFunc("/api/system/versions/", func(w http.ResponseWriter, r *http.Request){
		v := Versions{
			Linux: Uname(),
			Rkt: RktVer(),
		}

		js, err := json.Marshal(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	serveMux.HandleFunc("/api/system/vpn/", func (w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vpn := GetVPNService(c)
		if vpn == nil {
			w.Write([]byte("false"))
			return
		}

		if vpn.LoadState == "loaded" &&
		   vpn.ActiveState == "active" &&
		   vpn.SubState == "running" &&
		   vpn.PID != 0 {
			w.Write([]byte("true"))
			return
		}

		w.Write([]byte("false"))
	})

	serveMux.HandleFunc("/api/system/memory/", func (w http.ResponseWriter, r *http.Request){

		m := RAM()
		if m == nil {
			http.Error(w, "dunno", http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(m)
		if err != nil {
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
 * Gets local rkt version as a string
 */
func RktVer() string {
	out, err := exec.Command("rkt", "v").Output()
	if err != nil {
		return err.Error()
	}
	lines := bytes.Split(out, []byte{'\n'})
	fields := strings.Fields(string(lines[0]))
	return fields[2]
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
	m.Total, err = strconv.ParseInt(t[1], 10, 64)
	m.Avail, err = strconv.ParseInt(av[1], 10, 64)

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


/* Old disk info way -- only got mounted info, not unmounted blocks
type Disk struct {
	FS      string
	Mount   string
	All     uint64
	Used    uint64
	Free    uint64
	TInodes uint64
	FInodes uint64
}

func DiskUsage(path string) (*Disk, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return nil, err
	}
	d := Disk{
		Mount: path,
		All: fs.Blocks * uint64(fs.Bsize),
		Free: fs.Bavail * uint64(fs.Bsize),
		TInodes: fs.Files,
		FInodes: fs.Ffree,
	}
	d.Used = d.All - d.Free

	return &d, nil
}

type Mount struct {
	FS    string
	Mount string
	Type  string
}
func GetMounts() ([]Mount, error) {
	var m []Mount

	procMount, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(procMount), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)

		if !strings.HasPrefix(fields[0], "/") && fields[0] != "tmpfs" {
			continue
		}
		m = append(m, Mount{fields[0], fields[1], fields[2]})
	}

	return m, nil
}
*/




type BlockDevice struct {
	// populated from /proc/partitions
	maj          int
	min          int
	RawSize      uint64
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
		rem, err := ioutil.ReadFile("/sys/block/"+fields[3]+"/removable")
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

	b.All = fs.Blocks * uint64(fs.Bsize) - reserved
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
