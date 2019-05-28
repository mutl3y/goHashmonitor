package hashmonitor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var hasAdmin bool

// var supportedCards = []string{"Radeon Vega Frontier Edition", "Radeon RX 570 Series", "Radeon RX 580 Series", "Radeon RX Vega", "Radeon (TM) RX 470 Graphics", "Radeon (TM) RX 480 Graphics"}

func init() {
	switch OS := runtime.GOOS; {
	case OS == "windows":
		hasAdmin, _ = winElevationCheck()
		if !hasAdmin {
			log.Fatalf("You need to run this as admin or allow bypass of UAC")
		}
	case OS == "linux":

	default:
		log.Fatalf("OS not yet supported: %v ", OS)
	}

}

// winCmd run a command in a shell in windows and returns the output as a readWriter
// use powershell if you need to use or bypass UAC
func winCmd(path, command string) (io.ReadWriter, error) {
	var sysCmd bool
	in := strings.Fields(command)
	cm, a := in[0], in[1:]

	// check if supported system command
	builtin := []string{"cmd", "powershell"}
	for _, v := range builtin {
		if cm == v {
			sysCmd = true
		}
	}

	cmd := &exec.Cmd{}
	if !sysCmd {
		cm = "./" + cm
		cmd = exec.Command(cm, a...)
		cmd.Dir = path
	}

	if sysCmd {
		switch cm {
		case "cmd":
			sw, args := a[0], a[1:]
			argString := strings.Join(args, " ")
			argString = fmt.Sprintf("%v%v%v", path, pathSep, argString)
			cmd = exec.Command(cm, sw, argString)
		case "powershell":
			argString := ""
			if a[0] != "-Command" && a[0] != "-Help" {
				if path != "" {
					argString = path + pathSep
				}
			}

			// if a[0] == "-Command" {
			// 	args := strings.Join(a[1:], "")
			// 	a = []string{}
			// 	a = append(a, "-encodedCommand")
			// 	bstr := base64.StdEncoding.EncodeToString([]byte(args))
			// 	argString = bstr
			//
			// }

			argString += strings.Join(a, " ")

			cmd = exec.Command(cm, argString)
		default:
			return nil, fmt.Errorf("%v not implemented yet", cm)
		}

	}
	b := &bytes.Buffer{}
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return b, err
}
func winElevationCheck() (bool, error) {
	checkAdminPS := `& {([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] 'Administrator')}`

	command := fmt.Sprintf("powershell %s", checkAdminPS)

	reader, err := winCmd("", command)
	if err != nil {
		return false, errors.Wrap(err, "winElevationCheck winCmd ")
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, errors.Wrap(err, "winElevationCheck readAll ")
	}
	str := dropCR(data)
	b, err := strconv.ParseBool(fmt.Sprintf("%s", str))
	return b, errors.Wrap(err, "winElevationCheck parseBool ")
}

type devConCard struct {
	pcidev, name string
	running      bool
}

type CardData struct {
	cards        []devConCard
	dir          string
	resetEnabled bool
}

func NewCardData(c *viper.Viper) *CardData {
	d := &CardData{}
	// 	d.cards  =	make([]devConCard, 0, 20)
	d.dir = c.GetString("Core.Stak.Dir")
	d.resetEnabled = c.GetBool("Device.Reset.Enabled")
	return d
}

func (ca *CardData) String() string {
	var str, status string
	for _, v := range ca.cards {
		switch v.running {
		case true:
			status = "running"
		case false:
			status = "stopped"
		}
		str += fmt.Sprintf("%v, card %v, %v\n", v.name, status, v.pcidev)
	}

	return str
}
func (ca *CardData) GetStatus() error {
	// fmt.Printf("card config %v",ca)
	switch Os := runtime.GOOS; {
	case Os == "windows":

		by, err := winCmd(ca.dir, "devcon.exe status =display")
		if err != nil {
			log.Errorf("error reseting cards %v", err)
		}

		err = ca.devConParse(by)
		if err != nil {
			log.Errorf("error parsing devcon output:  %v", err)
		}
		return errors.Wrap(err, "failed updating device status")
	case Os == "linux":

		fmt.Println("device reset not available in Linux version")

		return nil
	default:
		return fmt.Errorf("%v not supported", Os)
	}

}
func (ca *CardData) ResetCards(force bool) error {
	if !ca.resetEnabled {
		return nil
	}

	switch Os := runtime.GOOS; {
	case Os == "windows":
		for _, v := range ca.cards {
			if !v.running || force {
				debug("Resetting %v", v.name)
				command := fmt.Sprintf("powershell devcon.exe restart \"@%v\"", v.pcidev)
				by, err := winCmd(ca.dir, command)
				if err != nil {
					return fmt.Errorf("error running %v", err)
				}

				err = ca.devConParse(by)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed resetting device %v", v.name))
				}
			}
		}
	default:
		return fmt.Errorf("%v not supported", Os)
	}
	err := ca.GetStatus()
	return err
}

func (ca *CardData) devConParse(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(devConSplitFunc)

	card := devConCard{}
	wholeCard := false

	ca.cards = make([]devConCard, 0, 20)

	for scanner.Scan() {
		s := scanner.Text()

		switch {
		case strings.HasPrefix(s, "PCI"):
			card.pcidev = s
		case strings.Contains(s, "Name:"):
			card.name = s[6:]
		case strings.Contains(s, "Driver is running"):
			card.running = true
			wholeCard = true
		case strings.Contains(s, "Device is disabled"):
			card.running = false
			wholeCard = true
		case strings.Contains(s, "Device is stopped"):
			card.running = false
			wholeCard = true
		default:
			debug("unreconized devcon response %v", s)
		}
		if wholeCard {
			ca.cards = append(ca.cards, card)
			card = devConCard{}

			wholeCard = false
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("invalid input: %s", err)
	}

	// if err := json.Unmarshal([]byte(js), mc); err != nil {
	// 	return fmt.Errorf("parsing error %s \n I managed to build %+v", err, js)
	// }

	return nil
}
func devConSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	str := string(data)

	if start := strings.Index(str, "PCI"); start >= 0 {
		if end := strings.Index(str, "\n"); end >= start {
			return end + 1, dropCR(data[start:end]), nil
		}
	}

	if start := strings.Index(str, "Name:"); start >= 0 {
		if end := strings.Index(str, "\n"); end >= start {
			return end + 1, dropCR(data[start:end]), nil
		}
	}
	if start := strings.Index(str, "Device"); start >= 0 {
		if end := strings.Index(str, "\n"); end >= start {
			return end + 1, dropCR(data[start:end]), nil
		}
	}
	if start := strings.Index(str, "Driver"); start >= 0 {
		if end := strings.Index(str, "\n"); end >= start {
			return end + 1, dropCR(data[start:end]), nil
		}
	}
	if start := strings.Index(str, "device(s)"); start >= 0 {
		if end := strings.Index(str, "\n"); end >= start {
			return 0, nil, nil
		}
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[0 : len(data)-1]
	}
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data = data[0 : len(data)-1]
	}

	return data
}
