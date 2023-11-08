package dedupe

import "plane.watch/lib/tracker/beast"

func makeBeastMessages(iterMax int) []*beast.Frame {
	//maxVal := 0x00FFFFFF
	maxVal := iterMax * iterMax * iterMax
	messages := make([]*beast.Frame, 0, maxVal)

	// setup our test data
	template := make([]byte, len(beastModeSShort))
	copy(template, beastModeSShort)
	template[13] = 0
	template[14] = 0
	template[15] = 0
	for x := 0; x <= iterMax; x++ {
		for y := 0; y <= iterMax; y++ {
			for z := 0; z <= iterMax; z++ {
				shrt := make([]byte, len(beastModeSShort))
				copy(shrt, template)
				shrt[13] = byte(x)
				shrt[14] = byte(y)
				shrt[15] = byte(z)
				frame, _ := beast.NewFrame(shrt, false)
				messages = append(messages, frame)
			}
		}
	}
	return messages
}
