package utils

import "github.com/micmonay/keybd_event"

func PressEnterKey() {
	// https://github.com/gdamore/tcell/issues/194
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return
	}
	//set keys
	kb.SetKeys(keybd_event.VK_SPACE)

	//launch
	kb.Launching()
}