package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"sort"
)

func getBeastlyMessages(c *cli.Context) error {
	incomingChan, err := incoming(c)
	if nil != err {
		return err
	}
	log.Info().Msg("Processing...")

	outData := make(map[string]string)
	var keys []string

	frameCounter := 0

	for frame := range incomingChan {
		switch bFrame := frame.(type) {
		case *mode_s.Frame:
			println("Please supply only beast content")
			continue
		case *beast.Frame:
			if err := bFrame.Decode(); nil != err {
				fmt.Printf("Failed to decode: %X, %s\n", bFrame.Raw(), err)
				continue
			}
			avr := bFrame.AvrFrame()
			if nil == avr {
				fmt.Printf("Failed to decode: %X\n", bFrame.Raw())
				continue
			}
			frameCounter++
			idx := fmt.Sprintf("DF%02d_MT%02d_ST%02d", avr.DownLinkType(), avr.MessageType(), avr.MessageSubType())
			//println(idx)
			if _, ok := outData[idx]; ok {
				continue
			}
			keys = append(keys, idx)
			binData := ""
			for _, val := range bFrame.Raw() {
				binData += fmt.Sprintf("0x%02X,", val)
			}
			//fmt.Println(idx, bFrame.String(), binData)

			outData[idx] = binData
		}

	}

	println("Found ", len(outData), " different frame types out of", frameCounter, "messages")

	sort.Strings(keys)

	fmt.Println("messages := map[string][]byte{")
	keyStr := ""
	for _, name := range keys {
		fmt.Println(`  "` + name + `": {` + outData[name] + `},`)

		keyStr += "`" + name + "`, "
	}
	fmt.Println("}")
	fmt.Println("keys := []string{", keyStr, "}")

	return nil
}
