package hashmonitor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

func (mc *AmdConf) Read(r io.ReadCloser) error {
	scanner := bufio.NewScanner(r)
	defer func() { _ = r.Close() }()
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

func (mc *AmdConf) Write(rwc io.WriteCloser) (err error) {
	jsTmp, err := json.MarshalIndent(mc, "", "    ")
	stakStyle := string(jsTmp)
	if err != nil {
		return
	}

	// remove curly's
	stakStyle = stakStyle[:len(stakStyle)-1]
	stakStyle = stakStyle[1:] + ","
	// stakStyle += ","

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
