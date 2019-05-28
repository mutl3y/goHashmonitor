package hashmonitor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// type algo struct {
// 	Intensity struct {
// 		Analyser  bool
// 		MinStable time.Time
// 		Precision bool
// 	}
// }

type GpuThreadsConf struct {
	Index        int  `json:"index"`
	Intensity    int  `json:"intensity"`
	Worksize     int  `json:"worksize"`
	AffineToCPU  bool `json:"affine_to_cpu"`
	StridedIndex int  `json:"strided_index"`
	MemChunk     int  `json:"mem_chunk"`
	Unroll       int  `json:"unroll"`
	CompMode     bool `json:"comp_mode"`
	Interleave   int  `json:"interleave"`
}

type AmdConf struct {
	GpuThreadsConf []GpuThreadsConf `json:"gpu_threads_conf"`
	AutoTune       int              `json:"auto_tune"`
	PlatformIndex  int              `json:"platform_index"`
}

func NewAmdConfig() AmdConf {
	return AmdConf{}
}

// "gpu_threads_conf" :
// [
//     { "index" : 0, "Intensity" : 1000, "worksize" : 8, "affine_to_cpu" : false,
//       "strided_index" : true, "mem_chunk" : 2, "unroll" : 8, "comp_mode" : true,
//       "interleave" : 40
//     },
// ],

//noinspection GoUnhandledErrorResult
func (mc *AmdConf) gpuConfParse(r io.ReadCloser) error {
	scanner := bufio.NewScanner(r)
	defer r.Close()
	scanner.Split(bufio.ScanLines)
	var js string
	var removeComments = func(s string) string {
		var rtnString string
		switch {
		case strings.HasPrefix(s, "/*"):
		case strings.HasPrefix(s, " *"):
		case strings.Contains(s, "//"):
		default:
			rtnString += s
		}

		return rtnString
	}

	for scanner.Scan() {
		s := scanner.Text()
		js += removeComments(s)
	}

	// encase in curly's
	js = "{" + js + "}"

	// remove trailing comma's
	js = strings.NewReplacer("},]", "}]", ",}", "}").Replace(js)

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("invalid input: %s", err)
	}

	if err := json.Unmarshal([]byte(js), mc); err != nil {
		return fmt.Errorf("parsing error %s \n I managed to build %+v", err, js)
	}

	return nil
}

// amdIntTemplate Generates Config files with Intensity from min to max for cycling through
func (mc *AmdConf) amdIntTemplate(interleave int, dir string) (str string, err error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 666)
		if err != nil {
			log.Fatalf("failed to create directory %v \t%v", dir, err)
		}

	}

	for i := range mc.GpuThreadsConf {
		mc.GpuThreadsConf[i].Interleave = interleave
	}
	jsTmp, err := json.Marshal(mc)
	stakStyle := string(jsTmp)
	if err != nil {
		return
	}

	// remove curly's
	stakStyle = stakStyle[:len(stakStyle)-1]
	stakStyle = stakStyle[1:]
	stakStyle += ","
	// str	:= fmt.Sprintf("",mc.)
	str = fmt.Sprintf("amd_%v.txt", interleave)

	f, err := os.OpenFile(dir+str, os.O_WRONLY, 0666)
	if err != nil {
		return "", fmt.Errorf("amdIntTemplate_open %v", err)
	}
	defer f.Close()
	if _, err = f.WriteString(stakStyle); err != nil {
		return "", fmt.Errorf("amdIntTemplate_write %v", err)
	}

	return
}

func (mc *AmdConf) Write(rwc io.WriteCloser) (err error) {
	jsTmp, err := json.Marshal(mc)
	stakStyle := string(jsTmp)
	if err != nil {
		return
	}

	// remove curly's
	stakStyle = stakStyle[:len(stakStyle)-1]
	stakStyle = stakStyle[1:]
	stakStyle += ","

	if _, err = rwc.Write([]byte(stakStyle)); err != nil {
		log.Fatalf("failed to write amd.txt")
	}

	return rwc.Close()

}

func (mc *AmdConf) Map() (m map[string]interface{}) {
	m = make(map[string]interface{})
	m["autoTune"] = mc.AutoTune
	m["platformIndex"] = mc.PlatformIndex
	for k, v := range mc.GpuThreadsConf {
		id := fmt.Sprintf("gpu_%v.thread_%v.", v.Index, k)
		m[id+"Intensity"] = v.Intensity
		m[id+"Worksize"] = v.Worksize
		m[id+"StridedIndex"] = v.StridedIndex
		m[id+"MemChunk"] = v.MemChunk
		m[id+"Unroll"] = v.Unroll
		m[id+"Interleave"] = v.Interleave

	}
	return
}
